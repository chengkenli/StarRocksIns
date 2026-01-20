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
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"time"
)

var (
	router = gin.New()
	mu     sync.RWMutex
)

func main1() {
	// 初始路由设置
	setupRoutes()

	// 重载API（固定路由，不会被重载影响）
	//router.Any("/admin/reload", func(c *gin.Context) {
	//	reloadRoutes()
	//	c.String(200, "路由已重载")
	//})

	router.Run(fmt.Sprintf(":%d", util.Read.Server.Port))
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
		firstPrefix := fmt.Sprintf("%s_%s_avgs", m["app"].(string), time.Now().Format("20060102"))

		once.Do(func() {
			router.GET("/", func(c *gin.Context) {
				c.HTML(http.StatusOK, "main.html", nil)
			})
		})

		app := m["app"].(string)
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
			router.GET(path, createHandler(firstPrefix+template))
		}
	}

	// 重新设置重载API（确保每次重载后都存在）
	//router.Any("/admin/reload", func(c *gin.Context) {
	//	reloadRoutes()
	//	c.String(200, "路由已重载")
	//})
}

func createHandler(templateName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, templateName, nil)
	}
}

func reloadRoutes() {
	util.Loggrs.Info("重载路由...")
	firstPrefix = fmt.Sprintf("%s_%s_avgs", util.P.App, time.Now().Format("20060102"))
	setupRoutes()
}
