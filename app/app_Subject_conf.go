/*
 *@author  chengkenli
 *@project StarRocksIns
 *@package app
 *@file    app_Subject_conf
 *@date    2026/1/8 14:27
 */

package app

import (
	"StarRocksIns/conn"
	"StarRocksIns/tools"
	"StarRocksIns/util"
	"encoding/xml"
	"fmt"
	"gorm.io/gorm"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// SubjectConfig
// 配置巡检
func SubjectConfig(app string, db *gorm.DB, fs, bs []map[string]interface{}) (string, int, int) {
	var all []string

	nodeSum := len(fs) + len(bs)
	// 检查内核
	body, diffSum, itemSum := confinsp(fs, bs)
	// 检查hadoop
	hadoop, diffSum2, itemSum2 := configres(fs, bs)
	// 生成结论
	score, grade, concl := confgeninsp(diffSum+diffSum2, nodeSum, itemSum+itemSum2)
	// 巡检结论
	all = append(all, bodyConclusion(conclusion{
		Score:      score,
		ScoreLevel: grade,
		ScoreConcl: concl,
	}))

	all = append(all, hadoop...)
	all = append(all, body...)

	tools.WriteFile(strings.NewReplacer("*", firstPrefix+"_conf").Replace(util.Read.Server.Loadhtmlglob),
		bodyTemplates(statictemplates{
			App:     app,
			Title:   "配置巡检",
			Version: srversion,
			Select:  strings.Join(selectids(), "\n"),
			Body:    strings.Join(all, "\n"),
		}))
	util.Loggrs.Info(app, "巡检完成。")
	return concl, score, diffSum + diffSum2
}

// 智能生成巡检结论
func confgeninsp(diffSum, nodeSum, itemSum int) (int, string, string) {
	var grade, conclusion string
	sc := float64(itemSum-diffSum) / float64(itemSum) * 100
	switch {
	case int(sc) > 99:
		grade = "卓越"
		conclusion = fmt.Sprintf("本次巡检内核参数共【<b>%d</b>】个，涉及巡检节点有【<b>%d</b>】个，集群所有参数一致。\n\n", itemSum, nodeSum)
	default:
		grade = "一般"
		conclusion = fmt.Sprintf("本次巡检内核参数共【<b>%d</b>】个，涉及巡检节点有【<b>%d</b>】个，发现集群有【<b>%d</b>】处参数出现异常，内核参数在不同节点间的运行时值不一致，这可能导致集群资源倾斜。需要核实并统一这些参数的配置！\n\n", itemSum, nodeSum, diffSum)
	}
	return int(sc), grade, conclusion
}

// hadoop配置文件巡检
func configres(fs, bs []map[string]interface{}) ([]string, int, int) {
	if util.Read.Server.Privateuser == "" || util.Read.Server.Privatekey == "" {
		return nil,-1,-1
	}
	var filename, body, confSum, diffSum []string
	// 登录机器
	ot := conn.ConnectSSH(
		&util.ConnSSH{
			User:           util.Read.Server.Privateuser,
			Host:           fs[0]["IP"].(string),
			Port:           22,
			PrivateKeyFile: util.Read.Server.Privatekey,
			Command:        fmt.Sprintf("ls -l %s*.xml", strings.NewReplacer("log", "conf").Replace(felogpath)),
		})
	text := strings.Split(ot.(string), "\n")
	for _, item := range text {
		filename = append(filename, findFields(item))
	}
	//先获取文件名
	var confs []variable
	var juctconf, initconf [][]string
	for _, name := range filename {
		if name == "" {
			continue
		}
		// fe检查hadoop
		for _, f := range fs {
			ot := conn.ConnectSSH(
				&util.ConnSSH{
					User:           util.Read.Server.Privateuser,
					Host:           f["IP"].(string),
					Port:           22,
					PrivateKeyFile: util.Read.Server.Privatekey,
					Command:        fmt.Sprintf("cat %s/%s", strings.NewReplacer("log", "conf").Replace(felogpath), name),
				})
			fmat := parseXMLPropertiesRobust(ot.(string))
			if fmat == nil {
				continue
			}
			confSum = append(confSum, fmat...)
			confs = append(confs, variable{
				IP:       f["IP"].(string),
				Variable: fmat,
			})
		}

		// be检查hadoop
		for _, b := range bs {
			ot := conn.ConnectSSH(
				&util.ConnSSH{
					User:           util.Read.Server.Privateuser,
					Host:           b["IP"].(string),
					Port:           22,
					PrivateKeyFile: util.Read.Server.Privatekey,
					Command:        fmt.Sprintf("cat %s/%s", strings.NewReplacer("log", "conf").Replace(belogpath), name),
				})
			fmat := parseXMLPropertiesRobust(ot.(string))
			if fmat == nil {
				continue
			}
			confSum = append(confSum, fmat...)
			confs = append(confs, variable{
				IP:       b["IP"].(string),
				Variable: fmat,
			})
		}
		// 这里进行过滤
		if confInData != nil {
			for _, item := range confInData {
				confs = rmSlicevari(confs, item["key"].(string))
			}
			//
		}

		for _, c := range confs {
			for _, s := range c.Variable {
				key := strings.Split(s, "=")
				var val string
				if key[0] != "" {
					val = key[1]
				} else {
					val = ""
				}
				initconf = append(initconf, []string{`<span class="status status-ok">指标</span>`, key[0], val})
			}
		}

		bbody := findInconsistentVariables2(confs)
		if bbody == nil {
			continue
		}

		for _, b := range bbody {
			juctconf = append(juctconf, []string{b.IP, name, b.Variable})
			diffSum = append(diffSum, b.Variable)
		}
	}

	if !util.P.Purify {
		body = append(body, bodySubject(subject{
			Title2:  "hadoop配置",
			Thead:   []string{"IP", "配置文件", "配置"},
			Tbody:   diagnMark(juctconf),
			Comment: `<h2 class="section-title"><p class="tooltip" data-tooltip="以下属于跨节点hadoop配置不一致的风险项">hadoop配置</p></h2>`,
		}))

		return body, len(tools.RmdupSlice(diffSum)), len(tools.RmdupSlice(confSum))
	} else {
		body = append(body, bodySubject(subject{
			Title2:  "hadoop配置巡检项",
			Thead:   []string{"状态", "参数", "值"},
			Tbody:   SliceUnique(initconf),
			Comment: `<h2 class="section-title"><p class="tooltip" data-tooltip="以下属于hadoop配置巡检项">hadoop配置巡检项</p></h2>`,
		}))

		body = append(body, bodySubject(subject{
			Title2:  "hadoop异常配置",
			Thead:   []string{"IP", "配置文件", "配置"},
			Tbody:   diagnMark(nil),
			Comment: `<h2 class="section-title"><p class="tooltip" data-tooltip="以下属于跨节点hadoop配置不一致的风险项">hadoop异常配置</p></h2>`,
		}))
		return body, 0, 0
	}
}

// 根据路径，获取文件名
func findFields(text string) string {
	// 分割字符串获取最后一部分
	parts := strings.Fields(text)
	if len(parts) > 0 {
		filePath := parts[len(parts)-1]
		fileName := filepath.Base(filePath)
		return fileName
	}
	return ""
}

// 更健壮的版本，处理XML可能没有根元素的情况
func parseXMLPropertiesRobust(xmlStr string) []string {
	type Property struct {
		Name  string `xml:"name"`
		Value string `xml:"value"`
	}
	type Configuration struct {
		Properties []Property `xml:"property"`
	}

	var config Configuration
	err := xml.Unmarshal([]byte(xmlStr), &config)
	if err != nil {
		//util.Loggrs.Error("XML解析失败: " + err.Error())
		return nil
	}

	var result []string
	for _, prop := range config.Properties {
		line := fmt.Sprintf("%s=%s", prop.Name, prop.Value)
		result = append(result, line)
	}
	return result
}

// 内核配置巡检
func confinsp(fs, bs []map[string]interface{}) ([]string, int, int) {
	var body, iplist []string
	var confsum int
	var confs []variable

	for _, f := range fs {
		iplist = append(iplist, f["IP"].(string))
	}
	for _, b := range bs {
		iplist = append(iplist, b["IP"].(string))
	}
	// sysctl内核参数配置巡检
	var feconf, beconf, initconf [][]string
	for _, ip := range iplist {
		// 登录机器
		ot := conn.ConnectSSH(
			&util.ConnSSH{
				User:           util.Read.Server.Privateuser,
				Host:           ip,
				Port:           22,
				PrivateKeyFile: util.Read.Server.Privatekey,
				Command:        "cat /etc/sysctl.conf | egrep -v '#|$^'",
			})
		// 设置日志内容
		//b := strings.NewReplacer("\n", "<br>").Replace(ot.(string))
		var once sync.Once
		var title string
		for _, item := range strings.Split(ot.(string), "\n") {
			if item == "" {
				continue
			}
			confsum = confsum + 1
			tag := strings.Split(item, "=")
			once.Do(func() {
				title = ip

			})
			feconf = append(feconf, []string{title, tag[0], tag[1]})
			initconf = append(initconf, []string{`<span class="status status-ok">指标</span>`, tag[0], tag[1]})
			title = ""
		}
		confs = append(confs, variable{
			IP:       ip,
			Variable: strings.Split(ot.(string), "\n"),
		})
	}

	// limits内核参数配置巡检
	for _, ip := range iplist {
		// 登录机器
		ot := conn.ConnectSSH(
			&util.ConnSSH{
				User:           util.Read.Server.Privateuser,
				Host:           ip,
				Port:           22,
				PrivateKeyFile: util.Read.Server.Privatekey,
				Command:        "cat /etc/security/limits.conf | egrep -v '#|$^'",
			})
		// 设置日志内容
		var once sync.Once
		var title string
		for _, item := range strings.Split(ot.(string), "\n") {
			if item == "" {
				continue
			}
			confsum = confsum + 1
			once.Do(func() {
				title = ip
			})
			// 编译正则表达式，匹配一个或多个连续空格
			regex := regexp.MustCompile(`\s+`).ReplaceAllString(item, " ")
			beconf = append(beconf, []string{title, regex})
			title = ""
		}
	}

	// 找到配置不一致的节点
	bbody := findInconsistentVariables2(confs)
	var juctconf [][]string
	for _, b := range bbody {
		juctconf = append(juctconf, []string{b.IP, b.Variable})
	}

	if !util.P.Purify {
		body = append(body, bodySubject(subject{
			Title2:  "sysctl内核参数",
			Thead:   []string{"IP", "内核参数"},
			Tbody:   diagnMark(juctconf),
			Comment: `<h2 class="section-title"><p class="tooltip" data-tooltip="以下属于跨节点内核参数不一致的风险项">sysctl内核参数</p></h2>`,
		}))
		bediff := findDifferences(beconf)
		body = append(body, bodySubject(subject{
			Title2:  "limits内核配置",
			Thead:   []string{"IP", "内核参数"},
			Tbody:   diagnMark(bediff),
			Comment: `<h2 class="section-title"><p class="tooltip" data-tooltip="以下属于starrocks专用的内核参数，/etc/sysctl.conf 与 /etc/security/limits.conf">limits内核配置</p></h2>`,
		}))
		return body, len(juctconf) + len(bediff), confsum
	} else {
		body = append(body, bodySubject(subject{
			Title2:  "内核参数巡检项",
			Thead:   []string{"状态", "参数", "值"},
			Tbody:   SliceUnique(initconf),
			Comment: `<h2 class="section-title"><p class="tooltip" data-tooltip="内核参数巡检项">内核参数巡检项</p></h2>`,
		}))

		body = append(body, bodySubject(subject{
			Title2:  "sysctl内核参数",
			Thead:   []string{"IP", "内核参数"},
			Tbody:   diagnMark(nil),
			Comment: `<h2 class="section-title"><p class="tooltip" data-tooltip="以下属于跨节点内核参数不一致的风险项">sysctl内核参数</p></h2>`,
		}))
		body = append(body, bodySubject(subject{
			Title2:  "limits内核配置",
			Thead:   []string{"IP", "内核参数"},
			Tbody:   diagnMark(nil),
			Comment: `<h2 class="section-title"><p class="tooltip" data-tooltip="以下属于starrocks专用的内核参数，/etc/sysctl.conf 与 /etc/security/limits.conf">limits内核配置</p></h2>`,
		}))
		return body, 0, confsum
	}
}
