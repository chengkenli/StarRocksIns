/*
 *@author  chengkenli
 *@project StarRocksIns
 *@package app
 *@file    gin_admin
 *@date    2026/1/14 9:53
 */

package app

import (
	"StarRocksIns/util"
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	router = gin.New()
	mu     sync.RWMutex
)

func initRoutes() {
	go initloadDateRoute()
	// 初始路由设置
	setupRoutes()
	err := router.Run(fmt.Sprintf(":%d", util.Read.Server.Port))
	if err != nil {
		util.Loggrs.Error(err.Error())
		return
	}
}

func setupRoutes() {
	mu.Lock()
	defer mu.Unlock()

	// 清空路由（新建实例）
	router = gin.New()

	// 基础中间件（每次重载都需要重新设置）
	store := cookie.NewStore([]byte("starrocks-insp"))
	router.Use(sessions.Sessions("visited", store))
	router.LoadHTMLGlob(util.Read.Server.Loadhtmlglob)
	router.Static("/static", util.Read.Server.Loadstatic)
	// 动态路由
	var once sync.Once
	for _, m := range util.MetaLink {
		app := m["app"].(string)

		firstPrefix := fmt.Sprintf("%s_%s", app, time.Now().Format("20060102"))

		once.Do(func() {
			router.GET("/", func(c *gin.Context) {
				util.Loggrs.Info(fmt.Sprintf("request: /, response: %s_index.html", firstPrefix))
				c.HTML(http.StatusOK, fmt.Sprintf("%s_index.html", firstPrefix), nil)
			})
		})

		routes := map[string]string{
			fmt.Sprintf("/%s", app):         "_index.html",
			fmt.Sprintf("/%s/log", app):     "_log.html",
			fmt.Sprintf("/%s/monitor", app): "_monitor.html",
			fmt.Sprintf("/%s/node", app):    "_node.html",
			fmt.Sprintf("/%s/avgs", app):    "_avgs.html",
			fmt.Sprintf("/%s/conf", app):    "_conf.html",
			fmt.Sprintf("/%s/olap", app):    "_olap.html",
		}

		for path, template := range routes {
			router.GET(path, createHandler("标准", path, firstPrefix+template))
		}
	}
}

func createHandler(model, path, templateName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		util.Loggrs.Info(fmt.Sprintf("%s request: %s, response: %s", model, path, templateName))
		c.HTML(http.StatusOK, templateName, nil)
	}
}

func reloadRoutes() {
	util.Loggrs.Info("重载路由...")
	setupRoutes()
}

func initloadDateRoute() {
	rewritefile()
	util.Loggrs.Info("注册历史路由。")
	loadhtmlglob := strings.NewReplacer("/*.html", "/history").Replace(util.Read.Server.Loadhtmlglob)
	// 清空路由（新建实例）
	r := gin.New()
	// 基础中间件（每次重载都需要重新设置）
	store := cookie.NewStore([]byte("starrocks-insp"))
	r.Use(sessions.Sessions("visited", store))
	r.LoadHTMLGlob(loadhtmlglob + "/*.html")
	r.Static("/static", util.Read.Server.Loadstatic)

	var availdate []availableDates
	dateitem := regexDates(findDir(strings.NewReplacer("/*.html", "").Replace(loadhtmlglob)))
	for _, m := range util.MetaLink {
		app := m["app"].(string)
		for _, date := range dateitem {
			firstPrefix := fmt.Sprintf("%s_%s", app, date)
			routes := map[string]string{
				fmt.Sprintf("/%s/%s", app, date):         "_index.html",
				fmt.Sprintf("/%s/%s/log", app, date):     "_log.html",
				fmt.Sprintf("/%s/%s/monitor", app, date): "_monitor.html",
				fmt.Sprintf("/%s/%s/node", app, date):    "_node.html",
				fmt.Sprintf("/%s/%s/avgs", app, date):    "_avgs.html",
				fmt.Sprintf("/%s/%s/conf", app, date):    "_conf.html",
				fmt.Sprintf("/%s/%s/olap", app, date):    "_olap.html",
			}
			for path, template := range routes {
				availdate = append(availdate,
					availableDates{
						Date: date,
						URL:  path,
					})
				r.GET(path, createHandler("历史", path, firstPrefix+template))
			}
		}
	}
	//availdateJson
	availdateJson, _ = json.Marshal(availdate)

	r.Run(fmt.Sprintf(":%d", util.Read.Server.Port+100))
}

// 迁移历史文件和重置接口
func rewritefile() {
	util.Loggrs.Info("reseting...")
	// 迁移历史文件
	loadhtmlglob := strings.NewReplacer("/*.html", "/history").Replace(util.Read.Server.Loadhtmlglob)
	os.MkdirAll(loadhtmlglob, 0755)
	files := findDir(strings.NewReplacer("/*.html", "").Replace(util.Read.Server.Loadhtmlglob))

	for _, file := range files {
		if !strings.Contains(file, time.Now().Format("20060102")) {
			oldfile := strings.NewReplacer("*.html", file).Replace(util.Read.Server.Loadhtmlglob)
			newfile := fmt.Sprintf("%s/%s", loadhtmlglob, filepath.Base(file))
			// 将文件迁移到历史目录
			err := os.Rename(oldfile, newfile)
			if err != nil {
				util.Loggrs.Error(err.Error())
				continue
			}
			// 获取集群名称
			app := strings.Split(file, "_")[0]
			// 获取文件日期标记
			matches := regexp.MustCompile(`_(\d{8})_`).FindStringSubmatch(file)
			if len(matches) > 1 {
				date := matches[1]
				replacehtml(newfile, getreplacements(app, date))
			}
		}
	}

	// 替换历史文件数据
	files = findDir(loadhtmlglob)
	for _, file := range files {
		newfile := fmt.Sprintf("%s/%s", loadhtmlglob, filepath.Base(file))

		if !strings.Contains(file, time.Now().Format("20060102")) {
			// 获取集群名称
			app := strings.Split(file, "_")[0]
			// 获取文件日期标记
			matches := regexp.MustCompile(`_(\d{8})_`).FindStringSubmatch(file)
			if len(matches) > 1 {
				date := matches[1]
				replacehtml(newfile, getreplacements(app, date))
			}
		}
	}
	util.Loggrs.Info("Ok，Ready...")
}
