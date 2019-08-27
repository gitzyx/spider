package main

import (
	"fmt"
	"time"

	"github.com/op/go-logging"
	//	"log/syslog"
	"os"
)

func init() {
	InitStdOutLog()
}

func SetLogLevel(logLevelStr string) error {
	var appLogLevel logging.Level
	if level, err := logging.LogLevel(logLevelStr); err != nil {
		appLogLevel = logging.DEBUG
	} else {
		appLogLevel = level
	}
	logging.SetLevel(appLogLevel, "")
	return nil
}

func MustGetLogger(module string) *logging.Logger {
	log := logging.MustGetLogger(module)
	return log
}

//将日志内容输出到可执行文件目录中的log文件中，方便本地调试
func InitStdOutLog() {

	os.Mkdir("./log/", 0666)
	var szLogFile string = "./log/" + os.Args[0] + "_" + time.Now().Format("2006-01-02") + ".log"
	logFile, err := os.OpenFile(szLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("open or create panic.log failure:", err)
	}

	backend2 := logging.NewLogBackend(logFile, "", 0)
	stdFormat := logging.MustStringFormatter("%{shortfile} %{longfunc}  %{message}")
	backendFormatter := logging.NewBackendFormatter(backend2, stdFormat)

	// Set the backends to be used.
	logging.SetBackend(backendFormatter) //syslog 和 stdout
}
