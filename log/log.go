package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// 文件输出等级，ERROR > WARNING > INFO > DEBUG
// 大于等于当前配置等级的日志，将被输出到文件
var Threshold = "INFO"

// 日志文件输出路径（文件夹地址）
var Dir = ""

// 日志输出配置初始化
const flag = log.Ldate | log.Ltime

var skip = 2

var debug = log.New(os.Stdout, "DEBUG   ", flag)
var info = log.New(os.Stdout, "INFO    ", flag)
var waring = log.New(os.Stdout, "WARNING ", flag)
var error = log.New(os.Stdout, "ERROR   ", flag)

var debugF = log.New(logTargetFile(), "DEBUG   ", flag)
var infoF = log.New(logTargetFile(), "INFO    ", flag)
var warningF = log.New(logTargetFile(), "WARNING ", flag)
var errorF = log.New(logTargetFile(), "ERROR   ", flag)

// Debug 输出调试级日志
func Debug(v ...interface{}) {
	debug.Println(output(v))
	if strings.ToLower(strings.TrimSpace(Threshold)) == "debug" {
		debugF.Println(output(v))
	}
}

// Info 输出信息级日志
func Info(v ...interface{}) {
	info.Println(output(v))
	threshold := strings.ToLower(strings.TrimSpace(Threshold))
	if threshold == "debug" || threshold == "info" {
		infoF.Println(output(v))
	}
}

// Warning 输出警告级日志
func Warning(v ...interface{}) {
	waring.Println(output(v))
	threshold := strings.ToLower(strings.TrimSpace(Threshold))
	if threshold == "debug" || threshold == "info" || threshold == "warning" {
		warningF.Println(output(v))
	}
}

// Error 输出错误级日志
func Error(v ...interface{}) {
	error.Println(output(v))
	errorF.Println(output(v))
}

func UseSetting() {
	debugF = log.New(logTargetFile(), "DEBUG   ", flag)
	infoF = log.New(logTargetFile(), "INFO    ", flag)
	warningF = log.New(logTargetFile(), "WARNING ", flag)
	errorF = log.New(logTargetFile(), "ERROR   ", flag)
}

func logTargetFile() io.Writer {
	date := time.Now().Format("2006-01-02")
	logFileName := ""
	if strings.TrimSpace(Dir) == "" {
		logFileName = "./" + date + ".log"
	} else {
		os.MkdirAll(Dir, os.ModePerm)
		if strings.LastIndex(Dir, "/") != len(Dir)-1 {
			Dir = Dir + "/"
		}
		logFileName = Dir + date + ".log"
	}
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	return logFile
}

func output(v []interface{}) string {
	fn := ""
	wd, _ := os.Getwd()
	_, p, l, _ := runtime.Caller(skip)
	if strings.Contains(p, wd) {
		fn = " " + p[len(wd):] + ":" + strconv.Itoa(l)
		return fmt.Sprintf(fn+" %s", v...)
	}
	for i := 0; i < 7; i++ {
		_, p, l, _ = runtime.Caller(i)
		if strings.Contains(p, wd) {
			fn = " " + p[len(wd):] + ":" + strconv.Itoa(l)
			skip = i
		}
	}
	return fmt.Sprintf(fn+" %s", v...)
}
