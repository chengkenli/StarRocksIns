/*
 *@author  chengkenli
 *@project StarRocksIns
 *@package app
 *@file    init
 *@date    2026/1/7 14:35
 */

package app

import (
	"StarRocksIns/conn"
	"StarRocksIns/util"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"gorm.io/gorm"
	"time"
)

var (
	monitorInData, logInData, avgsInData, confInData, olapInData []map[string]interface{}
	srversion                                                    string
	felogpath                                                    string
	belogpath                                                    string
	done                                                         chan struct{}
	faillog                                                      []string
	availdateJson                                                []byte
)

func init() {
	gin.SetMode(gin.ReleaseMode) // 必须先设置模式

	faillog = []string{" error ", " fail ", " exception "}
	done = make(chan struct{}, 1)
	time.Sleep(2)
	go func() {
		for _, m := range util.MetaLink {
			if m["app"].(string) == util.P.App {
				felogpath = m["fe_log_path"].(string)
				belogpath = m["be_log_path"].(string)
			}
		}
	}()
	mydb, err := conn.ConnectMySQL()
	if err != nil {
		util.Loggrs.Error(err.Error())
		return
	}
	/////////////////////////////////////
	/////////////////////////////////////
	//done = make(chan struct{})
	defer func() { done <- struct{}{} }()
	// 初始化读取一下数据表
	if util.Read.Subject.Monitor != "" {
		r := mydb.Raw(fmt.Sprintf("select * from %s where state=1 order by id", util.Read.Subject.Monitor)).Scan(&monitorInData)
		if r.Error != nil {
			util.Loggrs.Error(r.Error.Error())
			return
		}
	}
	if util.Read.Subject.Avgs != "" {
		r := mydb.Raw(fmt.Sprintf("select * from %s where state=1", util.Read.Subject.Avgs)).Scan(&avgsInData)
		if r.Error != nil {
			util.Loggrs.Error(r.Error.Error())
			return
		}
	}
	if util.Read.Subject.Log != "" {
		r := mydb.Raw(fmt.Sprintf("select * from %s where state=1", util.Read.Subject.Log)).Scan(&logInData)
		if r.Error != nil {
			util.Loggrs.Error(r.Error.Error())
			return
		}
	}
	if util.Read.Subject.Conf != "" {
		r := mydb.Raw(fmt.Sprintf("select * from %s where state=1", util.Read.Subject.Conf)).Scan(&confInData)
		if r.Error != nil {
			util.Loggrs.Error(r.Error.Error())
			return
		}
	}
	if util.Read.Subject.Olap != "" {
		r := mydb.Raw(fmt.Sprintf("select * from %s where state=1", util.Read.Subject.Olap)).Scan(&olapInData)
		if r.Error != nil {
			util.Loggrs.Error(r.Error.Error())
			return
		}
	}
	/////////////////////////////////////
	/////////////////////////////////////
}

// 获取fe和be的信息，并返回
// @avgs *gorm.DB 数据库连接对象
// @result []map[string]interface{}
func backend(db *gorm.DB) ([]map[string]interface{}, []map[string]interface{}) {
	var fenode []map[string]interface{}
	r := db.Raw("show frontends").Scan(&fenode)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return fenode, nil
	}
	var benode []map[string]interface{}
	r = db.Raw("show backends").Scan(&benode)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return fenode, benode
	}
	return fenode, benode
}

// 获取statistic信息
// @avgs *gorm.DB 数据库连接对象
// @result map[string]interface{}
func statistic(db *gorm.DB) (map[string]interface{}, []map[string]interface{}) {
	var m []map[string]interface{}
	r := db.Raw("show proc '/statistic'").Scan(&m)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return nil, m
	}
	for _, i2 := range m {
		if i2["DbId"].(string) == "Total" {
			return i2, m
		}
	}
	return nil, m
}

// 统计集群用户数
func statisuser(db *gorm.DB) int {
	var m []map[string]interface{}
	r := db.Raw("show users").Scan(&m)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return -1
	}
	return len(m)
}

// 统计每天集群访问量
func statisrequest(db *gorm.DB) int64 {
	var m map[string]interface{}
	r := db.Raw(fmt.Sprintf("select count(*) as count from %s where timestamp >= date_sub(now(), INTERVAL 24 HOUR) ", util.Read.Server.Auditlog)).Scan(&m)
	if r.Error != nil {
		util.Loggrs.Error(r.Error.Error())
		return -1
	}
	return m["count"].(int64)
}

type apistsatisNode struct {
	Data struct {
		Result []struct {
			ID       string      `json:"id"`
			ParentID interface{} `json:"parentId"`
			Name     string      `json:"name"`
			Env      string      `json:"env"`
			Type     string      `json:"type"`
			Node     string      `json:"node"`
			CPU      string      `json:"cpu"`
			Mem      string      `json:"mem"`
			Storage  string      `json:"storage"`
		} `json:"result"`
	} `json:"data"`
	Msg       string `json:"msg"`
	Code      int    `json:"code"`
	Timestamp int64  `json:"timestamp"`
}

// 统计集群明细数据
func statisnode() apistsatisNode {
	var aliasname string
	for _, m := range util.MetaLink {
		if m["app"].(string) == util.P.App {
			aliasname = m["alias"].(string)
		}
	}
	//发送POST请求并处理响应
	response, err := resty.New().
		R().
		SetHeaders(map[string]string{
			"Content-Type": "application/json",
		}).
		SetBody(fmt.Sprintf(`{"clusterId": "%s"}`, aliasname)).
		Post(util.Read.Server.Nodeurl)
	if err != nil {
		util.Loggrs.Warn("报错：", err.Error())
	}
	var r apistsatisNode
	json.Unmarshal(response.Body(), &r)
	return r
}
