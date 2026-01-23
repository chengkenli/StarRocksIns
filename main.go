/*
 *@author  chengkenli
 *@project StarRocksIns
 *@package StarRocksIns
 *@file    main
 *@date    2026/1/6 16:36
 */

package main

import (
	"StarRocksIns/app"
	_ "StarRocksIns/init"
	"StarRocksIns/util"
)

func main() {
	util.Init()
	app.App()
}
