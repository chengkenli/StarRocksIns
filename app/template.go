/*
 *@author  chengkenli
 *@project StarRocksIns
 *@package module
 *@file    module_template
 *@date    2026/1/7 16:38
 */

package app

import (
	"fmt"
	"strings"
	"time"
)

type statictemplates struct {
	App     string
	Title   string
	Version string
	Body    string
	Script  string
	Css     string
	Select  string
}

func bodyTemplates(t statictemplates) string {
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
	<script src="/static/chart.js"></script>
	<script src="/static/chartjs-adapter-date-fns.bundle.min.js"></script>
    <title>%s</title>
	<link rel="icon" type="image/x-icon" href="/static/favicon.ico" />
	%v
    <link rel="stylesheet" href="/static/node.css">
</head>
<body>
	<header>
        <div style="position: relative; margin-top: 20px;">
            <a href="/%s" style="position: absolute; left: 0; bottom: 0;color: white;">« 返回概览</a>
        </div>
    </header>

	<div class="watermark-container" id="watermark"></div>
    <div class="container">
        <header>
            <h1>%s</h1>
        </header>
        
        <div class="report-info">
            <div><strong>巡检日期：</strong>%s</div>
            <div><strong>集群版本：</strong>%s</div>
        </div>
        
		<!-- 集群选择标签页 -->
		<div style="text-align: right;">
			<div class="selector-container">
				<select id="appSelector">
					<option value="" selected disabled>选择集群</option>
					%v
				</select>
			</div>
		</div>

        <div class="content">
            %v
        </div>
        
        <footer>
            <p>StarRocks%s报告 | 巡检时间: %s</p>
            <p>本报告来自Infra Data - StarRocks</p>
        </footer>
    </div>

	<script>
        // 动态创建重复水印
        function createWatermark() {
            const container = document.getElementById('watermark');
            const text = 'StarRocks (%s)';
            const spacing = 200; // 水印间距
            
            for (let x = 0; x < window.innerWidth; x += spacing) {
                for (let y = 0; y < window.innerHeight; y += spacing) {
                    const watermark = document.createElement('div');
                    watermark.className = 'watermark';
                    watermark.textContent = text;
                    watermark.style.left = x + 'px';
                    watermark.style.top = y + 'px';
                    container.appendChild(watermark);
                }
            }
        }
        createWatermark();
    </script>
	<script>
        // 添加一些交互效果
        document.addEventListener('DOMContentLoaded', function() {
            const statusElements = document.querySelectorAll('.status');
            
            statusElements.forEach(status => {
                status.addEventListener('click', function() {
                    const statusText = this.textContent;
                });
            });
            
            // 添加按钮点击效果
            const buttons = document.querySelectorAll('.action-btn');
            buttons.forEach(button => {
                button.addEventListener('click', function() {
                    this.style.transform = 'scale(0.95)';
                    setTimeout(() => {
                        this.style.transform = '';
                    }, 150);
                });
            });
        });
    </script>
	<script>
        document.getElementById('appSelector').addEventListener('change', function() {
            const selectedValue = this.value;
            if (selectedValue) {
                window.location.href = selectedValue;
            }
        });
    </script>
	%v
</body>
</html>`, t.Title, t.Css, t.App, t.Title, time.Now().Format("2006-01-02"), t.Version, t.Select,
		t.Body, t.Title, time.Now().Format("2006-01-02 15:04:05"), t.App, t.Script)
	return body
}

type conclusion struct {
	Score      int    //总评分
	ScoreLevel string //评分等级，卓越，优秀，良好，一般，差
	ScoreConcl string //结论
}

// Conclusion
// 巡检结论
func bodyConclusion(c conclusion) string {
	return fmt.Sprintf(`<div class="section">
        <h2 class="section-title">巡检总结</h2>
        <div class="overview">
            <div class="score-card">
                <div class="score-circle">
                    <div class="score-value">%d</div>
                </div>
                <div class="score-label">%s</div>
            </div>
            <!-- 巡检总结放在评分后面 -->
            <div class="conclusion">
                <h2><i class="fas fa-clipboard-check"></i>巡检结论</h2>
                <p class="conclusion-text">%s</p>
            </div>
        </div>
    </div>`, c.Score, c.ScoreLevel, c.ScoreConcl)
}

type indicator struct {
	Title3   string      //标题h3
	Value    interface{} //value
	Subvalue string      //sub-value
}

// IndicatorLast
// 后执行
// 显示指标
func bodyIndicatorLast(title2, subindi string) string {
	b2 := fmt.Sprintf(`<div class="section">
        <h2 class="section-title">%s</h2>
        <div class="summary-cards">
			%v
        </div>
    </div>`, title2, subindi)
	return b2
}

// IndicatorFirst
// 先执行
// 显示指标
func bodyIndicatorFirst(s indicator) string {
	b1 := fmt.Sprintf(`<div class="card">
        <h3>%s</h3>
        <div class="value">%v</div>
        <div class="sub-value">%s</div>
    </div>`, s.Title3, s.Value, s.Subvalue)
	return b1
}

type subject struct {
	Title2  string
	Thead   []string
	Tbody   [][]string
	Comment string
}

// Subject
// 明细内容
func bodySubject(s subject) string {
	var title string
	if s.Comment != "" {
		title = s.Comment
	} else {
		title = fmt.Sprintf(`<h2 class="section-title">%s</h2>`, s.Title2)
	}
	return fmt.Sprintf(`<div class="section">
         %s
         <table>
             <thead>
                 %s
             </thead>
             <tbody>
                 %s
             </tbody>
         </table>
     </div>`, title, subject2Sh(s.Thead), subject2td(s.Tbody))
}

// Subject2Sh
// th
func subject2Sh(ths []string) string {
	var th2 []string
	for _, th := range ths {
		th2 = append(th2, fmt.Sprintf("<th>%s</th>", th))
	}
	return fmt.Sprintf("<tr>%s</tr>", strings.Join(th2, "\n"))
}

// Subject2td
// td
func subject2td(tds [][]string) string {
	var trs []string
	for _, td := range tds {
		var td2 []string
		for _, sd := range td {
			td2 = append(td2, fmt.Sprintf("<td>%s</td>", sd))
		}
		trs = append(trs, fmt.Sprintf("<tr>%s</tr>", strings.Join(td2, "\n")))
	}
	return strings.Join(trs, "\n")
}
