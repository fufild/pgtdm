package logging

import (
	"fmt"
	"log"
	"time"
)

const red uint8 = 31
const green uint8 = 32
const yellow uint8 = 33
const gray uint8 = 37
const white uint8 = 97

func Log(message string) {
	currentTime := time.Now()
	timestamp := currentTime.Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] %s\n", timestamp, message)
}

func Panic(message string, args ...interface{}) {
	logMessage("panic", message, args...)
	log.Panic(applyColor(red, message))
}

func FatalError(message string, args ...interface{}) {
	logMessage("error", message, args...)
	log.Fatal(applyColor(red, message))
}

func Error(message string, args ...interface{}) {
	logMessage("error", message, args...)
}

func Warn(message string, args ...interface{}) {
	logMessage("warning", message, args...)
}

func Info(message string, args ...interface{}) {
	logMessage("info", message, args...)
}

func Success(format string, args ...interface{}) {
	logMessage("success", format, args...)
}

func logMessage(level string, message string, args ...interface{}) {
	color := gray
	switch level {
	case "error":
		color = red
	case "warning":
		color = yellow
	case "info":
		color = white
	case "success":
		color = green
	}

	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}
	println(applyColor(color, message))
}

func applyColor(color uint8, message string) string {
	currentTime := time.Now()
	timestamp := currentTime.Format("2006-01-02 15:04:05")
	return fmt.Sprintf("\033[%dm [%s] %s \033[0m", color, timestamp, message)
}
