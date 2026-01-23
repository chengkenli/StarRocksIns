/*
 *@author  chengkenli
 *@project StarRocksIns
 *@package tools
 *@file    Object
 *@date    2026/1/6 17:42
 */

package tools

import (
	"StarRocksIns/util"
	"bufio"
	"crypto/rand"
	"math/big"
	"os"
)

// RandomPassWord 根据指定长度的生产高敏感度字符串
func RandomPassWord(n int) string {
	allowedChars := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		// 生成一个随机索引
		ri, err := rand.Int(rand.Reader, big.NewInt(int64(len(allowedChars))))
		if err != nil {
			return ""
		}
		// 使用随机索引获取一个字符
		b[i] = allowedChars[ri.Int64()]
	}

	return string(b)
}

// WriteFile 文件落地
func WriteFile(fname, msg string) {
	fileHandle, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		util.Loggrs.Error(err.Error())
		return
	}
	defer fileHandle.Close()
	// NewWriter 默认缓冲区大小是 4096
	// 需要使用自定义缓冲区的writer 使用 NewWriterSize()方法
	buf := bufio.NewWriterSize(fileHandle, len(msg))

	buf.WriteString(msg)

	err = buf.Flush()
	if err != nil {
		util.Loggrs.Error(err.Error())
		return
	}
}

// RmdupSlice /*数组去重*/
func RmdupSlice(strs []string) []string {
	result := []string{}
	tempMap := map[string]byte{} // 存放不重复字符串
	for _, e := range strs {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

// RmInSlice 从数组中移除某个元素
func RmInSlice(slice []string, element string) []string {
	// 创建一个新的切片来保存结果
	var result []string
	// 遍历原切片，并将不等于element的元素添加到结果切片中
	for _, v := range slice {
		if v != element {
			result = append(result, v)
		}
	}
	return result
}

// StrInSlice 检查数组中是否存在某个元素
func StrInSlice(list []string, str string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
