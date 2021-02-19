package judge

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/leoleoasd/EduOJBackend/app/request"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/suntt2019/EduOJJudger/api"
	"github.com/suntt2019/EduOJJudger/base"
	"github.com/suntt2019/Judger"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var testCaseLocks sync.Map

func getSeccompRuleName(language string) (rules string) {
	switch language {
	case "c":
		rules = "c_cpp"
	case "cpp":
		rules = "c_cpp"
	default:
		rules = "general"
	}
	return
}

var ErrBuildError = errors.New("build error")
var ErrTLE = errors.New("time limit exceeded")
var ErrMLE = errors.New("memory limit exceeded")
var ErrRTE = errors.New("runtime error")
var ErrDSC = errors.New("dangerous system call")
var ErrWA = errors.New("wrong answer")

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
		err := api.ErrNotAvailable
		for err == api.ErrNotAvailable {
			task, err = api.GetTask()
		}
		if err != nil {
			log.WithField("error", err).Error("Error occurred when getting task.")
		}
		err = api.UpdateRun(task.RunID, generateRequest(task, judge(task)))
		if err != nil {
			log.WithField("error", err).Error("Error occurred when sending update request.")
		}
	}
	base.QuitWG.Done()
}

func generateRequest(task *api.Task, judgementError error) *request.UpdateRunRequest {
	req := request.UpdateRunRequest{
		Status:             "",
		MemoryUsed:         task.MemoryUsed,
		TimeUsed:           task.TimeUsed,
		OutputStrippedHash: task.OutputStrippedHash,
		Message:            "",
	}
	//switch task.JudgeResult {
	//case judger.SUCCESS:
	//	if task.CompareResult {
	//		req.Status = "ACCEPTED"
	//	} else {
	//		req.Status = "WRONG_ANSWER"
	//	}
	//case judger.CPU_TIME_LIMIT_EXCEEDED:
	//	req.Status = "TIME_LIMIT_EXCEEDED"
	//case judger.REAL_TIME_LIMIT_EXCEEDED:
	//	req.Status = "TIME_LIMIT_EXCEEDED"
	//case judger.MEMORY_LIMIT_EXCEEDED:
	//	req.Status = "MEMORY_LIMIT_EXCEEDED"
	//case judger.RUNTIME_ERROR:
	//	req.Status = "RUNTIME_ERROR"
	//default:
	//	if judgementError != nil {
	//		judgementError = errors.New(fmt.Sprintf("unexpected running result: %d", task.JudgeResult))
	//	}
	//}
	if errors.Is(judgementError, ErrBuildError) {
		req.Status = "COMPILE_ERROR"
	} else if judgementError != nil {
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

	buildOutput, err := ioutil.TempFile("", "eduoj_judger_build_output_*")
	if err != nil {
		return errors.Wrap(err, "could not create temp file for build output")
	}
	task.BuildOutputPath = buildOutput.Name()
	if err := buildOutput.Close(); err != nil {
		return errors.Wrap(err, "could not close build output")
	}

	runFile, err := ioutil.TempFile("", "eduoj_judger_run_file_*")
	if err != nil {
		return errors.Wrap(err, "could not create temp file")
	}
	task.RunFilePath = runFile.Name()
	if err := runFile.Close(); err != nil {
		return errors.Wrap(err, "could not close run file")
	}

	compareOutput, err := ioutil.TempFile("", "eduoj_judger_compare_output_*")
	if err != nil {
		return errors.Wrap(err, "could not create temp file for compare output")
	}
	task.CompareOutputPath = compareOutput.Name()
	if err := buildOutput.Close(); err != nil {
		return errors.Wrap(err, "could not close compare output")
	}

	if err = api.GetFile(task.CodeFile, path.Join(task.JudgeDir, "code")); err != nil {
		return errors.Wrap(err, "could not get input file")
	}

	if err = Build(task); err != nil {
		return errors.Wrap(err, "could not build user program")
	}

	if err = Run(task); err != nil {
		return errors.Wrap(err, "could not run user program")
	}

	if err = hashOutput(task); err != nil {
		return errors.Wrap(err, "could not hash output")
	}

	//if task.CompareResult, err = compare(task); err != nil {
	//	return errors.Wrap(err, "could not compare output")
	//}

	return nil
}

func getTestCase(task *api.Task) error {
	l, _ := testCaseLocks.LoadOrStore(task.TestCaseID, &sync.Mutex{})
	l.(*sync.Mutex).Lock()
	defer l.(*sync.Mutex).Unlock()

	updatedAtPath := path.Join(viper.GetString("path.test_cases"), strconv.Itoa(int(task.TestCaseID)), "updated_at")
	if err := os.MkdirAll(path.Join(viper.GetString("path.test_cases"), strconv.Itoa(int(task.TestCaseID))), 0700); err != nil { // TODO:perm
		return errors.Wrap(err, "could not create directory for test cases")
	}
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
	updatedAt, err := os.Create(updatedAtPath)
	if err != nil {
		return errors.Wrap(err, "could not get updated_at file")
	}
	if err = updatedAt.Close(); err != nil {
		return errors.Wrap(err, "could not close updated_at file")
	}
	return nil
}

func Build(task *api.Task) error {
	var err error
	if err = exec.Command("chmod", "-R", "777", task.JudgeDir).Run(); err != nil {
		return errors.Wrap(err, "could not set permission for judge directory")
	}

	if err = EnsureLatestScript(task.Language.BuildScript.Name, task.Language.BuildScript.UpdatedAt); err != nil {
		return errors.Wrap(err, "could not ensure build script latest")
	}

	ctx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("judge.build.max_time"))
	defer cancel()
	cmd := exec.CommandContext(ctx, path.Join(viper.GetString("path.scripts"), task.Language.BuildScript.Name, "run"),
		append([]string{task.JudgeDir}, strings.Split(task.BuildArg, " ")...)...)

	buildOutput, err := os.OpenFile(task.BuildOutputPath, os.O_WRONLY, 0)
	if err != nil {
		return errors.Wrap(err, "could not open build output file")
	}
	defer buildOutput.Close()
	cmd.Stdout = buildOutput
	cmd.Stderr = buildOutput

	if err := base.BuildUser.Run(cmd); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return ErrBuildError
		}
		return errors.Wrap(err, "fail to build user program")
	}
	return nil
}

func Run(task *api.Task) error {

	var runScriptOutput string
	var err error
	if runScriptOutput, err = RunScriptWithOutput(task.Language.RunScript.Name, task.Language.RunScript.UpdatedAt, task.JudgeDir); err != nil {
		return errors.Wrap(err, "could not ensure run script latest")
	}
	RunCommand := strings.Split(runScriptOutput, " ")

	result, err := judger.Run(judger.Config{
		MaxCPUTime:           int(task.TimeLimit),
		MaxRealTime:          int(task.TimeLimit),
		MaxMemory:            int32(task.MemoryLimit),
		MaxStack:             int32(task.MemoryLimit),
		MaxProcessNumber:     -1,
		MaxOutputSize:        viper.GetInt32("judge.run.max_output_size"),
		MemoryLimitCheckOnly: 0,
		ExePath:              RunCommand[0],
		InputPath:            task.InputFilePath,
		OutputPath:           task.RunFilePath,
		ErrorPath:            os.DevNull,
		Args:                 RunCommand[1:],
		Env:                  []string{},
		LogPath:              viper.GetString("log.sandbox_log_path"),
		SeccompRuleName:      getSeccompRuleName(task.Language.Name),
		Uid:                  base.RunUser.Uid,
		Gid:                  base.RunUser.Gid,
	})

	if err != nil {
		return errors.Wrap(err, "fail to run user program")
	}

	task.TimeUsed = uint(result.CPUTime)
	task.MemoryUsed = uint(result.Memory)
	if syscall.Signal(result.Signal) == syscall.SIGSYS {
		return ErrDSC
	}

	switch result.Result {
	case judger.CPU_TIME_LIMIT_EXCEEDED:
		fallthrough
	case judger.REAL_TIME_LIMIT_EXCEEDED:
		return ErrTLE
	case judger.MEMORY_LIMIT_EXCEEDED:
		return ErrMLE
	case judger.RUNTIME_ERROR:
		return ErrRTE
	case judger.SYSTEM_ERROR:
		return errors.New("system error")
	}
	return nil
}

func hashOutput(task *api.Task) error {
	f, err := os.Open(task.RunFilePath)
	if err != nil {
		return errors.Wrap(err, "could not open run file")
	}
	defer f.Close()
	hh := sha256.New()
	_, err = io.Copy(hh, &base.StrippedReader{Inner: bufio.NewReader(f)})
	if err != nil {
		return errors.Wrap(err, "could not open run file")
	}
	task.OutputStrippedHash = hex.EncodeToString(hh.Sum(nil))
	return nil
}

func Compare(task *api.Task) error {
	err := EnsureLatestScript(task.CompareScript.Name, task.CompareScript.UpdatedAt)
	if err != nil {
		return errors.Wrap(err, "could not ensure compare script latest")
	}

	cmd := exec.Command("./run", task.RunFilePath, task.OutputFilePath, task.JudgeDir, task.InputFilePath)
	cmd.Dir = path.Join(viper.GetString("path.scripts"), task.CompareScript.Name)
	compareOutput, err := os.OpenFile(task.CompareOutputPath, os.O_WRONLY, 0)
	if err != nil {
		return errors.Wrap(err, "could not open compare output file")
	}
	defer compareOutput.Close()
	cmd.Stdout = compareOutput
	cmd.Stderr = compareOutput

	if err := cmd.Run(); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			if c := err.(*exec.ExitError).ExitCode(); c == 1 {
				return ErrWA
			} else {
				return errors.New(fmt.Sprintf("unexpected compare script output: %d", c))
			}
		}
		return errors.Wrap(err, "could not run compare script")
	}
	return nil
}
