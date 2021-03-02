package main

import (
	"github.com/EduOJ/judgeServer/judge"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os/user"
)

func main() {
	initConsoleLogger()
	u, err := user.Current()
	if err != nil {
		log.WithField("error", err).Fatal("could not get current user")
	}
	if u.Uid != "0" || u.Gid != "0" || u.Username != "root" {
		log.Fatal("Required root to run EduOJJudger")
	}
	readConfig()
	initFileLogger()
	initHttpClient()
	initUsers()

	judge.Start(viper.GetInt("thread_count"))
}
