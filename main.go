package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/suntt2019/EduOJJudger/judge"
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

	judge.Work(viper.GetInt("thread_count"))

	//task := api.Task{
	//	RunID:              0,
	//	Language:           models.Language{
	//		Name:             "",
	//		ExtensionAllowed: nil,
	//		BuildScriptName:  "",
	//		BuildScript:      &models.Script{
	//			Name:      "cpp_build",
	//			Filename:  "cpp_build",
	//			CreatedAt: time.Time{},
	//			UpdatedAt: time.Now(),
	//		},
	//		RunScriptName:    "",
	//		RunScript:        &models.Script{
	//			Name:      "cpp_run",
	//			Filename:  "cpp_run",
	//			CreatedAt: time.Time{},
	//			UpdatedAt: time.Time{},
	//		},
	//		CreatedAt:        time.Time{},
	//		UpdatedAt:        time.Time{},
	//	},
	//	TestCaseID:         0,
	//	InputFile:          "",
	//	OutputFile:         "",
	//	TestCaseUpdatedAt:  time.Time{},
	//	CodeFile:           "",
	//	InputFilePath:      "/home/sun123t2/mine/Project/EduOJ/EduOJJudger/test.in",
	//	OutputFilePath:     "/home/sun123t2/mine/Project/EduOJ/EduOJJudger/test.out",
	//	RunFilePath:        "/home/sun123t2/mine/Project/EduOJ/EduOJJudger/run_output.txt",
	//	BuildOutputPath:    "/home/sun123t2/mine/Project/EduOJ/EduOJJudger/build_output.txt",
	//	JudgeDir:           "/home/sun123t2/mine/Project/EduOJ/EduOJJudger/test_judge",
	//	MemoryLimit:        102400000,
	//	TimeLimit:          10000,
	//	BuildArg:           "",
	//	CompareScript:      models.Script{
	//		Name:      "test_cmp",
	//		Filename:  "test_cmp",
	//		CreatedAt: time.Time{},
	//		UpdatedAt: time.Time{},
	//	},
	//	TimeUsed:           0,
	//	MemoryUsed:         0,
	//	OutputStrippedHash: "",
	//	CompareOutputPath:  "compare_output.txt",
	//}
	////log.Error(judge.Build(&task))
	////log.Error(judge.Run(&task))
	//log.Error(judge.Compare(&task))
	//log.Errorf("%+v",task)
}
