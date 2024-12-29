package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	ERROR LogLevel = "ERROR"
)

type LoggerStruct struct {
	Logger      *log.Logger
	LogFile     *os.File
	environment string
}

// Permite definir el formato de los logs (en modo prod no se ve los logs DEBUG)
func (l *LoggerStruct) Log(message string, level LogLevel) {
	if l.environment == "prod" && level == "DEBUG" {
		return
	}

	format := fmt.Sprintf("[%s]: %s", level, message)
	l.Logger.Println(format)
}

func (l *LoggerStruct) CloseLogger() {
	l.LogFile.Close()
}

// Inicializo el log segun el archivo que le mande
func ConfigurarLogger(filename, env string) *LoggerStruct {
	logFile, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	logger := LoggerStruct{Logger: log.New(mw, "", log.LstdFlags), LogFile: logFile, environment: env}
	return &logger
}
