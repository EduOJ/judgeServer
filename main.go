package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/suntt2019/EduOJJudger/judge"
	"os/user"
	"time"
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

	// for test
	log.Error(judge.RunScript("test_script", time.Now()))
}
