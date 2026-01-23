/*
 *@author  chengkenli
 *@project StarRocksIns
 *@package app
 *@file    app_Subject_index
 *@date    2026/1/9 10:31
 */

package app

import (
	"StarRocksIns/tools"
	"StarRocksIns/util"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

// SubjectIndex
// 巡检首页
func SubjectIndex(app string, db *gorm.DB, fs, bs []map[string]interface{}) {
	initwriteDate()
	tic, fullTic := statistic(db)
	// 监控巡检【完成】
	_, score_monitor, diffSum_monitor := SubjectMonitor(app, db, fs, bs)
	// 配置巡检【完成】
	_, score_conf, diffSum_conf := SubjectConfig(app, db, fs, bs)
	// 节点巡检【完成】
	_, score_node, diffSum_node := SubjectNode(app, db, fs, bs, tic)
	// 内表巡检【完成】
	_, score_olap, diffSum_olap := SubjectOlap(app, db, fs, bs, tic, fullTic)
	// 日志巡检【完成】
	_, score_log, diffSum_log := SubjectLog(app, db, fs, bs)
	// 参数巡检【完成】
	_, score_avgs, diffSum_avgs := SubjectAvgs(app, db, fs, bs)
	////////////////////////////////////////////////////////
	score := (score_node + score_log + score_avgs + score_conf + score_olap + score_monitor) / 6
	var conclmsg []string
	if score_monitor < 100 {
		conclmsg = append(conclmsg, fmt.Sprintf("<b>监控巡检</b>"))
	}
	if score_log < 100 {
		conclmsg = append(conclmsg, fmt.Sprintf("<b>日志巡检</b>"))
	}
	if score_node < 100 {
		conclmsg = append(conclmsg, fmt.Sprintf("<b>机器巡检</b>"))
	}
	if score_avgs < 100 {
		conclmsg = append(conclmsg, fmt.Sprintf("<b>参数巡检</b>"))
	}
	if score_conf < 100 {
		conclmsg = append(conclmsg, fmt.Sprintf("<b>配置巡检</b>"))
	}
	if score_olap < 100 {
		conclmsg = append(conclmsg, fmt.Sprintf("<b>内表巡检</b>"))
	}
	ctxt := fmt.Sprintf("本次对StarRocks %s集群的巡检已完成，主要检查了集群节点状态、数据分片分布、查询性能指标、系统资源使用情况以及日志错误记录等关键内容。巡检结果显示，集群整体运行稳定，但其中在 %v 中发现潜在风险，需要进行核实处理。\n\n", app, strings.Join(conclmsg, ","))
	////////////////////////////////////////////////////////
	var all, indi []string
	// 获取集群明细信息
	if util.Read.Server.Nodeurl == "" {
		indi = append(indi, bodyIndicatorFirst(indicator{
			Title3:   "集群总节点数",
			Value:    len(fs) + len(bs),
			Subvalue: fmt.Sprintf("FE数：%d，BE数：%d", len(fs), len(bs)),
		}))
	} else {
		nodes := statisnode()
		for _, item := range nodes.Data.Result {
			if item.Type == "cluster" {
				indi = append(indi, bodyIndicatorFirst(indicator{
					Title3:   "集群总节点数",
					Value:    item.Node,
					Subvalue: fmt.Sprintf("FE数：%d，BE数：%d，总CPU：%s，总内存：%sgb，总存储：%sgb", len(fs), len(bs), item.CPU, item.Mem, item.Storage),
				}))
			}
		}
	}
	indi = append(indi, bodyIndicatorFirst(indicator{
		Title3:   "集群表数量",
		Value:    tic["TableNum"].(string),
		Subvalue: fmt.Sprintf("分区数量：%s，副本数：%s，分片数：%s", tic["PartitionNum"].(string), tic["TabletNum"].(string), tic["ReplicaNum"].(string)),
	}))
	u := statisuser(db)
	indi = append(indi, bodyIndicatorFirst(indicator{
		Title3:   "活跃用户数",
		Value:    u,
		Subvalue: "",
	}))
	sumrequest := statisrequest(db)
	indi = append(indi, bodyIndicatorFirst(indicator{
		Title3:   "24小时访问量",
		Value:    sumrequest,
		Subvalue: "",
	}))
	all = append(all, bodyIndicatorLast("集群规模", strings.Join(indi, "\n")))
	// 巡检结论
	var level string
	switch {
	case score > 99:
		level = "卓越"
	case score >= 90:
		level = "优秀"
	default:
		level = "一般"
	}
	all = append(all, bodyConclusion(conclusion{
		Score:      score,
		ScoreLevel: level,
		ScoreConcl: ctxt,
	}))
	// 巡检选项

	tb := [][]string{
		organize(organizes{
			app:     app,
			name:    "监控巡检",
			label:   "monitor",
			diffSum: diffSum_monitor,
			score:   score_monitor,
		}),
		organize(organizes{
			app:     app,
			name:    "日志巡检",
			label:   "log",
			diffSum: diffSum_log,
			score:   score_log,
		}),
		organize(organizes{
			app:     app,
			name:    "机器巡检",
			label:   "node",
			diffSum: diffSum_node,
			score:   score_node,
		}),
		organize(organizes{
			app:     app,
			name:    "参数巡检",
			label:   "avgs",
			diffSum: diffSum_avgs,
			score:   score_avgs,
		}),
		organize(organizes{
			app:     app,
			name:    "配置巡检",
			label:   "conf",
			diffSum: diffSum_conf,
			score:   score_conf,
		}),
		organize(organizes{
			app:     app,
			name:    "内表巡检",
			label:   "olap",
			diffSum: diffSum_olap,
			score:   score_olap,
		}),
	}
	all = append(all, bodySubject(subject{
		Title2:  "各模块巡检概览",
		Thead:   []string{"巡检模块", "摘要", "评分", "详情"},
		Tbody:   tb,
		Comment: `<h2 class="section-title"><p class="tooltip" data-tooltip="巡检总概览">各模块巡检概览</p></h2>`,
	}))
	// 写入文件
	tools.WriteFile(strings.NewReplacer("*", fmt.Sprintf("%s_%s", app, time.Now().Format("20060102"))+"_index").Replace(util.Read.Server.Loadhtmlglob),
		bodyTemplates(statictemplates{
			App:           app,
			Title:         fmt.Sprintf("%s巡检首页", strings.ToUpper(app)),
			Version:       srversion,
			Body:          strings.Join(all, "\n"),
			Select:        strings.Join(selectids(), "\n"),
			AvaildateJson: string(availdateJson),
			Watermark:     fmt.Sprintf("StarRocks（%s）", strings.ToUpper(app)),
		}))
	util.Loggrs.Info(app, "总结完成。")
}

// 整理结论信息
// @result 摘要，分数
type organizes struct {
	app, label, name string
	diffSum, score   int
}

func organize(o2 organizes) []string {
	var summ, score string
	// 生成摘要
	switch {
	case o2.score > 99:
		summ = fmt.Sprintf(`<p style="color: green;">检查通过，未发现明显问题。</p>`)
		score = fmt.Sprintf(`<p style="font-size: 20px; font-weight: bold; color: green;">%d</p>`, o2.score)
	default:
		switch o2.label {
		case "log":
			summ = fmt.Sprintf(`<p style="color: #ff8f00;">发现 <b>%d</b> 类日志相关风险，包括配置问题和错误记录。</p>`, o2.diffSum)
			if util.P.Purify {
				summ = summ + `<p style="color: green;font-weight: bold;">（已处理）</p>`
			}
		case "monitor":
			summ = fmt.Sprintf(`<p style="color: #ff8f00;">发现 <b>%d</b> 项监控指标风险，可能影响集群性能与稳定性。</p>`, o2.diffSum)
			if util.P.Purify {
				summ = summ + `<p style="color: green;font-weight: bold;">（已处理）</p>`
			}
		case "avgs":
			summ = fmt.Sprintf(`<p style="color: #ff8f00;">发现 <b>%d</b> 类参数配置风险，可能影响集群稳定性和可维护性。</p>`, o2.diffSum)
			if util.P.Purify {
				summ = summ + `<p style="color: green;font-weight: bold;">（已处理）</p>`
			}
		case "node":
			summ = fmt.Sprintf(`<p style="color: #ff8f00;">发现 <b>%d</b> 处操作系统参数与最佳实践不一致，可能影响集群性能和稳定性。</p>`, o2.diffSum)
			if util.P.Purify {
				summ = summ + `<p style="color: green;font-weight: bold;">（已处理）</p>`
			}
		case "conf":
			summ = fmt.Sprintf(`<p style="color: #ff8f00;">发现 <b>%d</b> 类远程文件相关风险，包括系统配置文件不一致和Hadoop配置问题。</p>`, o2.diffSum)
			if util.P.Purify {
				summ = summ + `<p style="color: green;font-weight: bold;">（已处理）</p>`
			}
		case "olap":
			summ = fmt.Sprintf(`<p style="color: #ff8f00;">共发现 <b>%d</b> 类 Tablet 相关风险，可能影响集群稳定性和性能。</p>`, o2.diffSum)
			if util.P.Purify {
				summ = summ + `<p style="color: green;font-weight: bold;">（已处理）</p>`
			}
		}
		score = fmt.Sprintf(`<p>%d</p>`, o2.score)
	}
	return []string{o2.name, summ, score, fmt.Sprintf(`<a href="/%s/%s">查看详情 »</a>`, o2.app, o2.label)}
}
