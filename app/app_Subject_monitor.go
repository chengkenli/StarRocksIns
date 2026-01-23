/*
 *@author  chengkenli
 *@project StarRocksInsp
 *@package app
 *@file    app_Subject_host
 *@date    2026/1/8 17:46
 */

package app

import (
	"StarRocksIns/tools"
	"StarRocksIns/util"
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
	"strings"
	"time"
)

// SubjectMonitor
// 机器巡检
func SubjectMonitor(app string, db *gorm.DB, fs, bs []map[string]interface{}) (string, int, int) {
	if monitorInData == nil {
		return "", 0, 0
	}
	var all []string
	// 监控巡检
	var oba []obtainData
	var initconf [][]string
	for _, item := range monitorInData {
		oba = append(oba, obtainData{
			Varname: item["expression"].(string),
			Title2:  item["name"].(string),
		})
		initconf = append(initconf, []string{""})
	}
	divs, scripts, diffvar, diffbody, okBody := obtain2line(oba)
	// 生成结论
	score, grade, concl := monitorgeninsp(diffvar, len(fs)+len(bs), len(monitorInData))
	// 巡检结论
	all = append(all, bodyConclusion(conclusion{
		Score:      score,
		ScoreLevel: grade,
		ScoreConcl: concl,
	}))

	// 生成列表
	all = append(all, bodySubject(subject{
		Title2: "重点监控指标巡检项",
		Thead:  []string{"状态", "名称", "当前值", "阈值"},
		Tbody:  append(diffbody, okBody...),
	}))
	all = append(all, divs...)

	tools.WriteFile(strings.NewReplacer("*", fmt.Sprintf("%s_%s", app, time.Now().Format("20060102"))+"_monitor").Replace(util.Read.Server.Loadhtmlglob),
		bodyTemplates(statictemplates{
			App:           app,
			Title:         "监控巡检",
			Version:       srversion,
			Body:          strings.Join(all, "\n"),
			Script:        strings.Join(scripts, "\n"),
			Select:        strings.Join(selectids(), "\n"),
			AvaildateJson: string(availdateJson),
			Watermark:     fmt.Sprintf("StarRocks（%s）", strings.ToUpper(app)),
			Css:           `<link rel="stylesheet" href="/static/ui-monitor.css">`,
		}))
	util.Loggrs.Info(app, "巡检完成。")
	return concl, score, len(diffvar)
}

// 智能生成巡检结论
func monitorgeninsp(diffvari []string, nodeSum, itemSum int) (int, string, string) {
	var grade, conclusion string
	sc := float64(itemSum-len(diffvari)) / float64(itemSum) * 100
	switch {
	case int(sc) > 99:
		grade = "卓越"
		conclusion = fmt.Sprintf("本次巡检监控指标共【<b>%d</b>】个，涉及巡检节点有【<b>%d</b>】个，集群所有监控指标皆正常。\n\n", itemSum, nodeSum)
	default:
		grade = "一般"
		var li []string
		for _, diff := range diffvari {
			for _, item := range monitorInData {
				if diff == item["key"].(string) {
					li = append(li, fmt.Sprintf(`<li>【%s】指标值≥【%d】</li>`, item["name"].(string), item["val"].(int64)))
				}
			}
		}
		conclusion = fmt.Sprintf("本次巡检监控指标共【<b>%d</b>】个，涉及巡检节点有【<b>%d</b>】个，"+
			"发现集群有【<b>%d</b>】处指标异常。需要核实并确认这些监控指标的是否正确！\n\n<ol>%v</ol>",
			itemSum,
			nodeSum,
			len(diffvari),
			strings.Join(li, "\n"))
	}
	return int(sc), grade, conclusion
}

type obtainData struct {
	Varname string
	Title2  string
	Comment string
}

// 解析监控数据，返回最大值
// 更简洁的版本，只返回最大值
func findSimpleMax(jsonData string) float64 {
	// 定义数据结构
	type PrometheusResponse struct {
		Status string `json:"status"`
		Data   struct {
			Result []struct {
				Metric struct {
					Name     string `json:"__name__"`
					Instance string `json:"instance"`
					Job      string `json:"job"`
					Type     string `json:"type"`
				} `json:"metric"`
				Values [][]interface{} `json:"values"`
			} `json:"result"`
			ResultType string `json:"resultType"`
		} `json:"data"`
	}

	var response PrometheusResponse
	err := json.Unmarshal([]byte(jsonData), &response)
	if err != nil {
		return 0
	}

	var maxValue float64

	for _, res := range response.Data.Result {
		for _, valuePair := range res.Values {
			if len(valuePair) != 2 {
				continue
			}

			var value float64
			switch v := valuePair[1].(type) {
			case string:
				fmt.Sscanf(v, "%f", &value)
			case float64:
				value = v
			}
			if value > maxValue {
				maxValue = value
			}
		}
	}
	return maxValue
}

// 根据监控指标获取数据并生成html元素
// @avgs 监控指标
// @return divs,scripts
func obtain2line(o []obtainData) ([]string, []string, []string, [][]string, [][]string) {
	var diffBody, okBody [][]string
	var c util.ConnectParms
	for _, m := range util.MetaLink {
		if m["app"].(string) == util.P.App {
			c = util.ConnectParms{
				Uri:        m["address"].(string),
				AaccessKey: m["manager_access_key"].(string),
				SecretKey:  m["manager_secret_key"].(string),
			}
		}
	}

	var divs, scripts, diffvar []string
	for _, item := range o {
		unix := time.Now().Unix()
		data := getOpenApi(
			util.Openapi{
				Sunix:      unix - 600,
				Eunix:      unix,
				Step:       10,
				Timeout:    10,
				Label:      item.Varname,
				Uri:        c.Uri,
				AaccessKey: c.AaccessKey,
				SecretKey:  c.SecretKey,
			},
		)

		var maxItem float64
		if !util.P.Purify {
			maxItem = findSimpleMax(string(data))
		} else {
			maxItem = findSimpleMax(string(data)) / 2
		}

		var comment string
		for _, oid := range monitorInData {
			if strings.Contains(item.Varname, oid["key"].(string)) {
				comment, diffvar = genComment(oid["key"].(string), diffvar, int64(maxItem), oid["val"].(int64))
				if strings.Contains(comment, "待优化") {
					// []string{"名称", "当前值", "阈值"}
					diffBody = append(diffBody, []string{`<span class="status status-warning">待确认</span>`, oid["name"].(string),
						fmt.Sprintf("%.2f", maxItem),
						fmt.Sprintf("%d", oid["val"].(int64)),
					})

					key := tools.RandomPassWord(10)
					div, script := script2Body2(string(data), item.Title2, comment, key)
					divs = append(divs, div)
					scripts = append(scripts, script)
				} else {
					okBody = append(okBody, []string{`<span class="status status-ok">正常</span>`, oid["name"].(string), "", fmt.Sprintf("%d", oid["val"].(int64))})
				}
			}
		}
	}
	return divs, scripts, diffvar, diffBody, okBody
}

// 根据监控信息，判断是否正常
// @avgs val:监控指标，key:监控值，maxItem：最大值，val：阈值指标，diffSum：差异累计
// @return 备注，差异累计数
func genComment(key string, diffvar []string, maxItem, value int64) (string, []string) {
	comment := `<span class="status status-ok">正常</span>`
	if maxItem >= value {
		comment = `<span class="status status-error">待优化</span>`
		diffvar = append(diffvar, key)
	}
	return comment, diffvar
}

// 生成html结构体
// @avgs json数据,h2标题,小字提示,图表名称
// @return div,script
func script2Body2(jds, h2, info, key string) (string, string) {
	div := fmt.Sprintf(`
<div class="container">
        <h2>%s</h2>
        <div class="info">%s</div>
        <div class="chart-container">
            <canvas id="Chart%s"></canvas>
        </div>
</div>`, h2, info, key)

	script := fmt.Sprintf(`
<script>
        // 解析Prometheus数据
        const prometheus_%s = %v;

        // 智能颜色生成器（支持任意数量节点）
        function generateColors(count) {
            const baseColors = ['#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA726', '#AB47BC', 
                               '#26C6DA', '#D4E157', '#5C6BC0', '#FF7043', '#78909C'];
            
            if (count <= baseColors.length) {
                return baseColors.slice(0, count);
            }
            
            // 如果节点数量超过基础颜色，生成随机颜色
            const colors = [...baseColors];
            for (let i = baseColors.length; i < count; i++) {
                colors.push('#' + Math.floor(Math.random()*16777215).toString(16));
            }
            return colors;
        }

        // 格式化字节数为更易读的格式
        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        // 处理数据
        const datasets_%s = [];
        const nodeCount_%s = prometheus_%s.data.result.length;
        const colors_%s = generateColors(nodeCount_%s);
        
        prometheus_%s.data.result.forEach((result, index) => {
            const instance = result.metric.instance;
            const values = result.values;
            
            const data = values.map(item => {
                const timestamp = item[0];
                const bytesUsed = parseInt(item[1]);
                
                // 使用格式化后的时间字符串作为x轴值
                const date = new Date(timestamp * 1000);
                const timeString = date.toLocaleTimeString('zh-CN', {
                    hour: '2-digit',
                    minute: '2-digit',
                    second: '2-digit'
                });
                return {
                    x: timeString, // 使用格式化时间
                    y: bytesUsed
                };
            });
            
            datasets_%s.push({
                label: instance,
                data: data,
                borderColor: colors_%s[index],
                backgroundColor: colors_%s[index] + '20',
                borderWidth: 2,
                pointRadius: 2,
                tension: 0.1,
                fill: false
            });
        });

        // 创建图表
        const ctx_%s = document.getElementById('Chart%s').getContext('2d');
        new Chart(ctx_%s, {
            type: 'line',
            data: {
                datasets: datasets_%s // 修正：应该是datasets而不是datasets_%s
            },
            options: {
                responsive: true,
                maintainAspectRatio: false,
                interaction: {
                    mode: 'index',
                    intersect: false
                },
                scales: {
                    x: {
                        type: 'category', // 使用分类轴显示格式化时间
                        title: {
                            display: true,
                            text: '时间'
                        },
                        ticks: {
                            maxTicksLimit: 10, // 限制显示的刻度数量
                            callback: function(value) {
                                return value; // 直接返回格式化后的时间字符串
                            }
                        }
                    },
                    y: {
                        title: {
                            display: true,
                            text: '使用量|使用频率|次数'
                        },
                        grid: {
                            color: 'rgba(0,0,0,0.1)'
                        },
                        ticks: {
                            callback: function(value) {
                                return formatBytes(value); // 格式化字节显示
                            }
                        }
                    }
                },
                plugins: {
                    legend: {
                        display: true,
                        position: 'top',
                        labels: {
                            usePointStyle: true,
                            padding: 15
                        }
                    },
                    tooltip: {
                        mode: 'index',
                        intersect: false,
                        backgroundColor: 'rgba(0,0,0,0.7)',
                        titleColor: '#fff',
                        bodyColor: '#fff',
                        callbacks: {
                            title: function(tooltipItems) {
                                return '时间: ' + tooltipItems[0].label;
                            },
                            label: function(context) {
                                const value = context.parsed.y;
                                return context.dataset.label + ': ' + formatBytes(value);
                            }
                        }
                    }
                }
            }
        });
	</script>`, key, jds, key, key, key, key, key, key, key, key, key, key, key, key, key, key)
	return div, script
}

func getOpenApi(avg util.Openapi) []byte {
	nonce := tools.RandomPassWord(32)
	//unix := time.Now().Unix()

	AuthCredential := fmt.Sprintf("%s/%d/%s", avg.AaccessKey, avg.Eunix, nonce)
	AuthCredentialEncrypted := util.HexEncrypt(AuthCredential, avg.SecretKey)
	body := fmt.Sprintf("start=%d&end=%d&step=%d&timeout=%d&query=%s", avg.Sunix, avg.Eunix, avg.Step, avg.Timeout, avg.Label)
	AuthContent := fmt.Sprintf(`HTTPMethod:POST
CanonicalURI:/openapi/v1/metrics/query-range
CanonicalQueryString:
CanonicalForm:%s`, body)
	AuthSignature := util.HexEncrypt(AuthContent, AuthCredentialEncrypted)
	AuthHeader := fmt.Sprintf("MANAGER-HMAC-SHA256 Credential=%s,Signature=%s", AuthCredential, AuthSignature)
	uri := fmt.Sprintf("%s/openapi/v1/metrics/query-range", avg.Uri)
	//发送POST请求并处理响应
	respones, err := resty.New().
		R().
		SetHeaders(map[string]string{
			"Content-Type":  "application/x-www-form-urlencoded",
			"Authorization": AuthHeader,
		}).
		SetBody(body).
		Post(uri)
	if err != nil {
		util.Loggrs.Warn("报错：", err.Error())
	}
	return respones.Body()
}
