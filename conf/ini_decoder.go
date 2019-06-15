package conf

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Load 加载指定路径中的全部INI文件
func Load(dir string) (map[string]string, error) {
	confs := make(map[string]string)
	// 遍历当前程序运行路径，读取所有INI文件
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	if strings.Index(dir, "/") != 0 {
		if dir != "" && dir != "." && dir != "./" {
			dir = "/" + dir
		}
	}
	locals := []string{"", ".", "./", "/"}
	for _, local := range locals {
		if dir != local {
			wd = wd + dir
			break
		}
	}
	filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		lowCaseName := strings.TrimSpace(strings.ToLower(info.Name()))
		if !strings.HasSuffix(lowCaseName, ".ini") {
			return nil
		}
		// 解析当前配置文件
		err = parseFile(path, confs)
		return err
	})
	return confs, nil
}

// LoadFile 加载指定路径的配置文件
func LoadFile(path string) (map[string]string, error) {
	confs := make(map[string]string)
	if err := parseFile(path, confs); err != nil {
		return nil, err
	}
	return confs, nil
}

// 解析指定路径的文件
// 将解析结果保存至confs，使用key、value存储
func parseFile(path string, confs map[string]string) error {
	section := ""
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	lineNum := 0
	for {
		lineBytes, _, err := r.ReadLine()
		lineNum++
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		line := strings.TrimSpace(string(lineBytes))
		// 判断注释定义
		if strings.Index(line, "#") == 0 || strings.Index(line, ";") == 0 {
			continue
		}
		// 判断节定义
		iLeft := strings.Index(line, "[")
		iRight := strings.Index(line, "]")
		iSharp := strings.Index(line, "#")
		iSemi := strings.Index(line, ";")
		iSlash := strings.Index(line, "//")
		if iLeft > -1 && iRight > -1 && iRight > iLeft+1 {
			if (iSharp > -1 && iSharp < iRight) || (iSemi > -1 && iSemi < iRight) || (iSlash > -1 && iSlash < iRight) {
				panic("initialization file syntax error: line " + strconv.Itoa(lineNum))
			}
			section = strings.TrimSpace(line[iLeft+1 : iRight])
			continue
		}
		// 判断key = value配置定义
		iEqual := strings.Index(line, "=")
		if iEqual < 1 {
			continue
		}
		key := strings.TrimSpace(line[:iEqual])
		val := strings.TrimSpace(line[iEqual+1:])

		if len(key) < 1 || strings.Contains(key, "#") || strings.Contains(key, ";") || strings.Contains(key, "//") {
			continue
		}
		if strings.Index(val, "#") == 0 || strings.Index(val, ";") == 0 || strings.Index(val, "//") == 0 {
			continue
		}
		if strings.Contains(val, "#") {
			val = val[0:strings.Index(val, "#")]
		}
		if strings.Contains(val, ";") {
			val = val[0:strings.Index(val, ";")]
		}
		if strings.Contains(val, "//") {
			val = val[0:strings.Index(val, "//")]
		}
		if section != "" {
			key = section + "." + key
		}
		confs[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}
	return nil
}
