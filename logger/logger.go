package logger

import "log"

var isDebug bool = false

func Info(format string, v ...interface{}) {
	log.SetPrefix("[INFO] ")
	log.Printf(format, v...)
}

func Debug(format string, v ...interface{}) {
	if isDebug {
		log.SetPrefix("[DBUG] ")
		log.Printf(format, v...)
	}
}

func Warn(format string, v ...interface{}) {
	log.SetPrefix("[WARN] ")
	log.Printf(format, v...)
}

func Error(format string, v ...interface{}) {
	log.SetPrefix("[ERRO] ")
	log.Printf(format, v...)
}

func OpenDebug() {
	isDebug = true
}

func init() {
	log.SetFlags(log.LstdFlags)
}
