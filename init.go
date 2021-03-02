package main

import (
	"bufio"
	"fmt"
	"github.com/EduOJ/judgeServer/base"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/writer"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func readConfig() {
	// TODO: set default
	log.Debug("Reading config.")
	viper.AddConfigPath(".")
	viper.SetConfigType("yml")
	if err := viper.ReadInConfig(); err != nil {
		log.WithField("error", err).Fatal("Could not read config.")
	}
	scriptPath := viper.GetString("path.scripts")
	if scriptPath[len(scriptPath)-1] == '/' {
		viper.Set("path.scripts", scriptPath[:len(scriptPath)-1])
	}
	runPath := viper.GetString("path.test_cases")
	if runPath[len(runPath)-1] == '/' {
		viper.Set("path.test_cases", runPath[:len(runPath)-1])
	}
	tempPath := viper.GetString("path.temp")
	if tempPath[len(tempPath)-1] == '/' {
		viper.Set("path.temp", tempPath[:len(tempPath)-1])
	}
}

func initConsoleLogger() {
	log.Debug("Initializing console logger.")
	log.SetOutput(os.Stdout)
	log.SetLevel(log.TraceLevel)

	_, fileName, _, _ := runtime.Caller(0)
	prefixPath := filepath.Dir(fileName)

	log.SetFormatter(&log.TextFormatter{
		PadLevelText:    true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05 MST",
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			function = f.Function + "()]"
			file = f.File
			if strings.HasPrefix(file, prefixPath) {
				file = file[len(prefixPath):]
			}
			file = fmt.Sprintf(" [.%s:%d", file, f.Line)
			return
		},
	})
	log.SetReportCaller(true)
}

func initFileLogger() {
	log.Debug("Initializing file logger.")
	filePath := viper.GetString("log.log_path")
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.WithField("error", err).Error("Failed to open log file")
	}

	log.AddHook(&writer.Hook{
		Writer: bufio.NewWriter(file),
		LogLevels: []log.Level{
			log.ErrorLevel,
			log.FatalLevel,
			log.PanicLevel,
		},
	})
}

func initHttpClient() {
	log.Debug("Initializing http client.")
	base.HttpClient = resty.New()
	base.HttpClient.SetHeader("Authorization", viper.GetString("auth.token")).
		SetHeader("Judger-Name", viper.GetString("auth.name")).
		SetHostURL(viper.GetString("backend.endpoint"))
}

func initUsers() {
	var err error
	err = base.BuildUser.Init(viper.GetString("user.build"))
	if err != nil {
		log.Fatal("Could not find compile user named " + viper.GetString("user.build"))
	}
	err = base.RunUser.Init(viper.GetString("user.run"))
	if err != nil {
		log.Fatal("Could not find run user named " + viper.GetString("user.run"))
	}
}
