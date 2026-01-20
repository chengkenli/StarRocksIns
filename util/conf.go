/*
 *@author  chengkenli
 *@project StarRocksQueris
 *@package util
 *@file    conf
 *@date    2024/8/7 14:42
 */

package util

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/viper"
	"net"
	"os"
	"path/filepath"
)

func usage() {
	fmt.Printf("\nUsage: %s [-s adhoc] [-h]\n\nOptions:\nStarRocks 巡检报告.\n", filepath.Base(os.Args[0]))
	flag.PrintDefaults()
	fmt.Println()
}

func init() {
	var (
		conf        string
		defaultConf string
	)

	path, err := findProcessPath(filepath.Base(os.Args[0]))
	if err != nil {
		execDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		defaultConf = fmt.Sprintf("%s/.%s.yaml", execDir, filepath.Base(os.Args[0]))
	}
	if path == "" {
		execDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		defaultConf = fmt.Sprintf("%s/.%s.yaml", execDir, filepath.Base(os.Args[0]))
	} else {
		defaultConf = fmt.Sprintf("%s/.%s.yaml", path, filepath.Base(os.Args[0]))
	}

	flag.StringVar(&conf, "c", defaultConf, "conf file")
	flag.BoolVar(&P.Help, "h", false, "show help information")
	flag.StringVar(&P.App, "s", "", "app")
	flag.BoolVar(&P.Agent, "agent", false, "只运行agent")
	flag.BoolVar(&P.Service, "service", false, "只运行service")
	flag.BoolVar(&P.Purify, "purify", false, "净化模式")

	flag.Parse()
	flag.Usage = usage

	if P.Help {
		flag.Usage()
		os.Exit(-1)
	}

	paths, name := filepath.Split(conf)
	Config := viper.New()
	Config.SetConfigFile(fmt.Sprintf("%s%s", paths, name))
	if err := Config.ReadInConfig(); err != nil {
		fmt.Println(err.Error())
	}
	// 读取配置文件所有内容放入内存
	marshal, err := json.Marshal(Config.AllSettings())
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	err = json.Unmarshal(marshal, &Read)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	// log init
	Logrus()
	// auth metadb
	Instglobal.vaild <- struct{}{}
	// send begin channel
	Instglobal.Begin <- struct{}{}
}

func Init() {
	/*获取当前主机的IP*/
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err.Error())
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				H.Ip = ipnet.IP.String()
			}
		}
	}
}

func findProcessPath(name string) (string, error) {
	processes, _ := process.Processes()
	for _, p := range processes {
		n, _ := p.Name()
		if n == name {
			exePath, err := p.Exe()
			if err != nil {
				return "", fmt.Errorf("failed to get exe path: %v", err)
			}
			return filepath.Dir(exePath), nil // 返回目录部分
		}
	}
	return "", fmt.Errorf("process not found")
}
