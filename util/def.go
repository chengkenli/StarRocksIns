/*
 *@author  chengkenli
 *@project StarRocksQueris
 *@package util
 *@file    def
 *@date    2024/8/7 14:44
 */

package util

import (
	"github.com/sirupsen/logrus"
)

var (
	Loggrs   *logrus.Logger
	P        ArvgParms
	MetaLink []map[string]interface{}
	H        Hosts
	Read     ReadInConf
)

type CustomLogger struct{}

func (l *CustomLogger) Errorf(format string, v ...interface{}) {}
func (l *CustomLogger) Warnf(format string, v ...interface{})  {} // 忽略 WARN
func (l *CustomLogger) Debugf(format string, v ...interface{}) {}

type ArvgParms struct {
	App     string
	Module  string
	Help    bool
	Agent   bool
	Service bool
	Purify  bool
}

type Hosts struct {
	Ip string
}
type ConnectParms struct {
	Host       string
	Port       int
	User       string
	Pass       string
	Base       string
	Uri        string
	AaccessKey string
	SecretKey  string
}

type Openapi struct {
	Sunix, Eunix, Step, Timeout int64
	Label                       string
	Uri                         string
	AaccessKey                  string
	SecretKey                   string
}

type ReadInConf struct {
	Metadb struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Base     string `json:"base"`
	} `json:"metadb"`
	Server struct {
		Port         int    `json:"port"`
		Loadhtmlglob string `json:"loadhtmlglob"`
		Loadstatic   string `json:"loadstatic"`
		Privateuser string `json:"privateuser"`
		Privatekey string `json:"privatekey"`
		Auditlog string `json:"auditlog"`
		Nodeurl string `json:"nodeurl"`
	} `json:"server"`
	Subject struct {
		Log     string `json:"log"`
		Monitor string `json:"monitor"`
		Avgs    string `json:"avgs"`
		Conf    string `json:"conf"`
		Olap    string `json:"olap"`
	} `json:"subject"`
	Log struct {
		Path string `json:"path"`
	} `json:"log"`
}

type ConnSSH struct {
	User           string
	Host           string
	Port           int
	PrivateKeyFile string
	Command        string
}
