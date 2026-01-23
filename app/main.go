/*
 *@author  chengkenli
 *@project StarRocksIns
 *@package app
 *@file    applica
 *@date    2026/1/6 16:37
 */

package app

import (
	"StarRocksIns/conn"
	"StarRocksIns/util"
)

func App() {
	for {
		select {
		case <-done:
			util.Loggrs.Info("数据库配置加载完成。")
			goto todo
		}
	}
todo:

	if util.P.Service || util.P.App == "" {
		util.Loggrs.Info("注册标准路由。")
		//go Fsnotify()
		initRoutes()
		return
	}

	app := util.P.App

	db, err := conn.StarRocks(app)
	if err != nil {
		util.Loggrs.Error(err.Error())
		return
	}
	var m map[string]interface{}
	db.Raw("select current_version() as version").Scan(&m)
	srversion = m["version"].(string)
	fs, bs := backend(db)
	// 首页
	SubjectIndex(app, db, fs, bs)
	if util.P.Agent {
		util.Loggrs.Info(app, "完成巡检，不做路由设置。")
		return
	}
	// web服务
	//go Fsnotify()
	initRoutes()
}
