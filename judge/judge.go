package judge

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"github.com/leoleoasd/EduOJBackend/app/request"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/suntt2019/EduOJJudger/api"
	"github.com/suntt2019/EduOJJudger/base"
	"github.com/suntt2019/Judger"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var testCaseLocks sync.Map
var languageToSeccompRuleName = map[string]string{
	"C":   "c_cpp",
	"C++": "c_cpp",
}

func Work(threadCount int) {
	base.QuitWG.Add(threadCount)
	for i := 0; i < threadCount; i++ {
		go work()
	}
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	<-s
	go func() {
		<-s
		log.Error("Force quitting")
		os.Exit(-1)
	}()

	log.Error("Server closing.")
	log.Error("Hit ctrl+C again to force quit.")
	base.Close()
	base.QuitWG.Wait()
}

func work() {
	stop := false
	go func() {
		<-base.BaseContext.Done()
		stop = true
	}()
	for !stop {
		var task *api.Task
		err := base.ErrNotAvailable
		for err == base.ErrNotAvailable {
			task, err = api.GetTask()
		}
		if err != nil {
			log.WithField("error", err).Error("Error occurred when getting task.")
		}
		err = api.UpdateRun(task.RunID, getRequest(task, judge(task)))
		if err != nil {
			log.WithField("error", err).Error("Error occurred when sending update request.")
		}
	}
	base.QuitWG.Done()
}

func getRequest(task *api.Task, judgementError error) *request.UpdateRunRequest {
	req := request.UpdateRunRequest{
		Status:             "",
		MemoryUsed:         task.MemoryUsed,
		TimeUsed:           task.TimeUsed,
		OutputStrippedHash: task.OutputStrippedHash,
		Message:            "",
	}
	switch task.JudgeResult {
	case judger.SUCCESS:
		if task.CompareResult {
			req.Status = "ACCEPTED"
		} else {
			req.Status = "WRONG_ANSWER"
		}
	case judger.CPU_TIME_LIMIT_EXCEEDED:
		req.Status = "TIME_LIMIT_EXCEEDED"
	case judger.REAL_TIME_LIMIT_EXCEEDED:
		req.Status = "TIME_LIMIT_EXCEEDED"
	case judger.MEMORY_LIMIT_EXCEEDED:
		req.Status = "MEMORY_LIMIT_EXCEEDED"
	case judger.RUNTIME_ERROR:
		req.Status = "RUNTIME_ERROR"
	default:
		if judgementError != nil {
			judgementError = errors.New(fmt.Sprintf("unexpected running result: %d", task.JudgeResult))
		}
	}
	if judgementError != nil {
		req.Status = "JUDGEMENT_FAILED"
		req.Message = judgementError.Error()
	}
	return &req
}

func judge(task *api.Task) error {
	var err error
	if err = getTestCase(task); err != nil {
		return errors.Wrap(err, "could not get test case")
	}

	if task.JudgeDir, err = ioutil.TempDir("", "eduoj_judger_run_*"); err != nil {
		return errors.Wrap(err, "could not create temp directory")
	}

	if err = api.GetFile(task.CodeFile, path.Join(task.JudgeDir, "code")); err != nil {
		return errors.Wrap(err, "could not get input file")
	}

	if err = build(task); err != nil {
		return errors.Wrap(err, "could not build user program")
	}

	if err = run(task); err != nil {
		return errors.Wrap(err, "could not run user program")
	}

	outBytes, err := ioutil.ReadAll(&base.StrippedReader{Inner: bufio.NewReader(task.RunFile)})
	h := sha256.Sum256(outBytes)
	task.OutputStrippedHash = string(h[:])

	if task.CompareResult, err = compare(task); err != nil {
		return errors.Wrap(err, "could not compare output")
	}

	return nil
}

func getTestCase(task *api.Task) error {
	l, _ := testCaseLocks.LoadOrStore(task.TestCaseID, &sync.Mutex{})
	l.(*sync.Mutex).Lock()
	defer l.(*sync.Mutex).Unlock()

	updatedAtPath := path.Join(viper.GetString("test_cases"), strconv.Itoa(int(task.TestCaseID)), "updated_at")
	ok, err := base.IsFileLatest(updatedAtPath, task.TestCaseUpdatedAt)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	if err = api.GetFile(task.InputFile, task.InputFilePath); err != nil {
		return errors.Wrap(err, "could not get input file")
	}
	if err = api.GetFile(task.OutputFile, task.OutputFilePath); err != nil {
		return errors.Wrap(err, "could not get output file")
	}
	if _, err = os.Create(updatedAtPath); err != nil {
		return errors.Wrap(err, "could not get updated_at file")
	}
	return nil
}

func build(task *api.Task) error {
	var err error
	if err = base.BuildUser.OwnMod(task.JudgeDir, 0600); err != nil {
		return errors.Wrap(err, "could not set permission for judge directory")
	}

	if err = base.BuildUser.OwnMod(path.Join(task.JudgeDir, "code"), 0400); err != nil {
		return errors.Wrap(err, "could not set permission for code")
	}

	if err = EnsureLatestScript(task.Language.BuildScript.Name, task.Language.BuildScript.UpdatedAt); err != nil {
		return errors.Wrap(err, "could not ensure build script latest")
	}

	result, err := judger.Run(judger.Config{
		MaxCPUTime:           viper.GetInt("judge.build.max_cpu_time"),
		MaxRealTime:          viper.GetInt("judge.build.max_real_time"),
		MaxMemory:            viper.GetInt32("judge.build.max_memory"),
		MaxStack:             viper.GetInt32("judge.build.max_stack"),
		MaxProcessNumber:     -1,
		MaxOutputSize:        -1,
		MemoryLimitCheckOnly: 0,
		ExePath:              path.Join(viper.GetString("path.scripts"), task.Language.BuildScript.Name),
		InputPath:            "",
		OutputPath:           "",
		ErrorPath:            "",
		Args: append([]string{
			path.Join(task.JudgeDir, "code"),
		}, strings.Split(task.BuildArg, ",")...),
		Env:             nil,
		LogPath:         viper.GetString("log.sandbox_log_path"),
		SeccompRuleName: "general",
		Uid:             base.BuildUser.Uid,
		Gid:             base.BuildUser.Gid,
	})
	if err != nil {
		return errors.Wrap(err, "fail to build user program")
	}
	if result.ExitCode != 0 {
		return errors.New(fmt.Sprintf("fail to build user program, build script returns %d", result.ExitCode))
	}
	return nil
}

func run(task *api.Task) error {
	var err error
	task.RunFile, err = ioutil.TempFile("", "eduoj_judger_run_file_*")
	if err != nil {
		return errors.Wrap(err, "could not create temp file")
	}

	if err = base.BuildUser.OwnModDir(task.JudgeDir, 0700); err != nil {
		return errors.Wrap(err, "could not set permission for judge directory")
	}

	if err = base.RunUser.OwnMod(task.InputFilePath, 0400); err != nil {
		return errors.Wrap(err, "could not set permission for input file")
	}

	if err = base.RunUser.OwnMod(task.RunFile.Name(), 0600); err != nil {
		return errors.Wrap(err, "could not set permission for run file")
	}

	result, err := judger.Run(judger.Config{
		MaxCPUTime:           int(task.TimeLimit),
		MaxRealTime:          int(task.TimeLimit), // TODO: real time
		MaxMemory:            int32(task.MemoryLimit),
		MaxStack:             int32(task.MemoryLimit), // TODO: Stack size
		MaxProcessNumber:     -1,
		MaxOutputSize:        viper.GetInt32("judge.run.max_output_size"),
		MemoryLimitCheckOnly: 0,
		ExePath:              path.Join(task.JudgeDir, "code"),
		InputPath:            task.InputFilePath,
		OutputPath:           task.OutputFilePath,
		ErrorPath:            "",
		Args:                 nil,
		Env:                  nil,
		LogPath:              viper.GetString("log.sandbox_log_path"),
		SeccompRuleName:      languageToSeccompRuleName[task.Language.Name],
		Uid:                  base.BuildUser.Uid,
		Gid:                  base.BuildUser.Gid,
	})
	if err != nil {
		return errors.Wrap(err, "fail to run user program")
	}

	task.TimeUsed = uint(result.CPUTime) // TODO: real time
	task.MemoryUsed = uint(result.Memory)
	task.JudgeResult = result.Result

	return nil
}

func compare(task *api.Task) (accepted bool, err error) {
	out, err := RunScriptWithOutput(task.CompareScript.Name, task.CompareScript.UpdatedAt)
	if err != nil {
		return false, errors.Wrap(err, "could not ensure latest compare script "+task.CompareScript.Name)
	}
	switch out {
	case "ACCEPTED":
		return true, nil
	case "WRONG_ANSWER":
		return false, nil
	default:
		return false, errors.New("unexpected compare script output: " + out)
	}
}
