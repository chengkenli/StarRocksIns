/*
 *@author  chengkenli
 *@project StarRocksIns
 *@package app
 *@file    fsnotifies
 *@date    2026/1/14 10:01
 */

package app

import (
	"StarRocksIns/util"
	"github.com/fsnotify/fsnotify"
	"strings"
)

func Fsnotify() {
	// 创建文件监视器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		util.Loggrs.Error(err.Error())
		return
	}
	defer watcher.Close()

	// 添加要监控的子目录
	dir := strings.NewReplacer("/*.html", "").Replace(util.Read.Server.Loadhtmlglob)
	err = watcher.Add(dir)
	if err != nil {
		util.Loggrs.Error(err.Error())
		return
	}

	// 处理文件变化事件
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Write == fsnotify.Write {
				util.Loggrs.Info("文件发生变化，重新注册路由！")
				reloadRoutes()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			util.Loggrs.Error(err.Error())
		}
	}
}
