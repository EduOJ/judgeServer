package judge

import (
	"encoding/hex"
	"fmt"
	"github.com/leoleoasd/EduOJBackend/app/request"
	"github.com/leoleoasd/EduOJBackend/database/models"
	"github.com/minio/sha256-simd"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/suntt2019/EduOJJudger/api"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"testing"
	"time"
)

func TestGetSeccompRuleName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "c_cpp", getSeccompRuleName("c"))
	assert.Equal(t, "c_cpp", getSeccompRuleName("cpp"))
	assert.Equal(t, "general", getSeccompRuleName("other_language"))
}

func TestGenerateRequest(t *testing.T) {
	t.Parallel()

	task := api.Task{
		TimeUsed:           3000,
		MemoryUsed:         5120,
		OutputStrippedHash: "test_generate_request_output_stripped_hash",
	}

	t.Run("Accepted", func(t *testing.T) {
		t.Parallel()
		expectedReq := request.UpdateRunRequest{
			Status:             "ACCEPTED",
			MemoryUsed:         5120,
			TimeUsed:           3000,
			OutputStrippedHash: "test_generate_request_output_stripped_hash",
			Message:            "",
		}
		actualReq := generateRequest(&task, nil)
		assert.Equal(t, &expectedReq, actualReq)
	})
	t.Run("WrongAnswer", func(t *testing.T) {
		t.Parallel()
		expectedReq := request.UpdateRunRequest{
			Status:             "WRONG_ANSWER",
			MemoryUsed:         5120,
			TimeUsed:           3000,
			OutputStrippedHash: "test_generate_request_output_stripped_hash",
			Message:            "",
		}
		actualReq := generateRequest(&task, errors.Wrap(ErrWA, "wrap message"))
		assert.Equal(t, &expectedReq, actualReq)
	})
	t.Run("PresentationError", func(t *testing.T) {
		t.Parallel()
		expectedReq := request.UpdateRunRequest{
			Status:             "PRESENTATION_ERROR",
			MemoryUsed:         5120,
			TimeUsed:           3000,
			OutputStrippedHash: "test_generate_request_output_stripped_hash",
			Message:            "",
		}
		actualReq := generateRequest(&task, errors.Wrap(ErrPE, "wrap message"))
		assert.Equal(t, &expectedReq, actualReq)
	})
	t.Run("TimeLimitExceeded", func(t *testing.T) {
		t.Parallel()
		expectedReq := request.UpdateRunRequest{
			Status:             "TIME_LIMIT_EXCEEDED",
			MemoryUsed:         5120,
			TimeUsed:           3000,
			OutputStrippedHash: "test_generate_request_output_stripped_hash",
			Message:            "",
		}
		actualReq := generateRequest(&task, errors.Wrap(ErrTLE, "wrap message"))
		assert.Equal(t, &expectedReq, actualReq)
	})
	t.Run("MemoryLimitExceeded", func(t *testing.T) {
		t.Parallel()
		expectedReq := request.UpdateRunRequest{
			Status:             "MEMORY_LIMIT_EXCEEDED",
			MemoryUsed:         5120,
			TimeUsed:           3000,
			OutputStrippedHash: "test_generate_request_output_stripped_hash",
			Message:            "",
		}
		actualReq := generateRequest(&task, errors.Wrap(ErrMLE, "wrap message"))
		assert.Equal(t, &expectedReq, actualReq)
	})
	t.Run("RuntimeError", func(t *testing.T) {
		t.Parallel()
		expectedReq := request.UpdateRunRequest{
			Status:             "RUNTIME_ERROR",
			MemoryUsed:         5120,
			TimeUsed:           3000,
			OutputStrippedHash: "test_generate_request_output_stripped_hash",
			Message:            "",
		}
		actualReq := generateRequest(&task, errors.Wrap(ErrRTE, "wrap message"))
		assert.Equal(t, &expectedReq, actualReq)
	})
	t.Run("DangerousSystemCalls", func(t *testing.T) {
		t.Parallel()
		expectedReq := request.UpdateRunRequest{
			Status:             "DANGEROUS_SYSTEM_CALLS",
			MemoryUsed:         5120,
			TimeUsed:           3000,
			OutputStrippedHash: "test_generate_request_output_stripped_hash",
			Message:            "",
		}
		actualReq := generateRequest(&task, errors.Wrap(ErrDSC, "wrap message"))
		assert.Equal(t, &expectedReq, actualReq)
	})
	t.Run("CompileError", func(t *testing.T) {
		t.Parallel()
		expectedReq := request.UpdateRunRequest{
			Status:             "COMPILE_ERROR",
			MemoryUsed:         5120,
			TimeUsed:           3000,
			OutputStrippedHash: "test_generate_request_output_stripped_hash",
			Message:            "",
		}
		actualReq := generateRequest(&task, errors.Wrap(ErrBuildError, "wrap message"))
		assert.Equal(t, &expectedReq, actualReq)
	})
	t.Run("JudgementFailed", func(t *testing.T) {
		t.Parallel()
		expectedReq := request.UpdateRunRequest{
			Status:             "JUDGEMENT_FAILED",
			MemoryUsed:         5120,
			TimeUsed:           3000,
			OutputStrippedHash: "test_generate_request_output_stripped_hash",
			Message:            "wrap message: other error",
		}
		actualReq := generateRequest(&task, errors.Wrap(errors.New("other error"), "wrap message"))
		assert.Equal(t, &expectedReq, actualReq)
	})
}

func TestCreateTempFiles(t *testing.T) {
	t.Parallel()

	task := api.Task{}
	err := createTempFiles(&task)
	assert.NoError(t, err)

	paths := map[string]string{
		"build_output":   task.BuildOutputPath,
		"run_file":       task.RunFilePath,
		"compare_output": task.CompareOutputPath,
	}

	for name, p := range paths {
		assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(`eduoj_judger_%s_\d+`, name)), p)
		f, err := os.OpenFile(p, os.O_WRONLY, 0)
		assert.NoError(t, err)
		_, err = f.WriteString("test_create_temp_files_write_string")
		assert.NoError(t, err)
		err = f.Close()
		assert.NoError(t, err)
		ff, err := os.Open(p)
		assert.NoError(t, err)
		b, err := ioutil.ReadAll(ff)
		assert.NoError(t, err)
		assert.Equal(t, "test_create_temp_files_write_string", string(b))
		err = ff.Close()
		assert.NoError(t, err)
	}
}

func TestGetTestCase(t *testing.T) {
	t.Parallel()

	t.Run("NewDownload", func(t *testing.T) {
		t.Parallel()

		id := hashStringToId("[TestCase] TestGetTestCase/NewDownload")

		latestUpdatedAt := time.Now()

		task := api.Task{
			TestCaseID:        id,
			InputFile:         "fileURI/test_get_test_case_new_download_input/test_get_test_case_new_download_input_content",
			OutputFile:        "fileURI/test_get_test_case_new_download_output/test_get_test_case_new_download_output_content",
			TestCaseUpdatedAt: latestUpdatedAt,
		}
		task.GenerateFilePath()
		err := getTestCase(&task)
		assert.NoError(t, err)

		checkFile(t, path.Join(viper.GetString("path.test_cases"), strconv.Itoa(int(id)), "in"), "test_get_test_case_new_download_input_content")
		checkFile(t, path.Join(viper.GetString("path.test_cases"), strconv.Itoa(int(id)), "out"), "test_get_test_case_new_download_output_content")
		checkFile(t, path.Join(viper.GetString("path.test_cases"), strconv.Itoa(int(id)), "updated_at"), "")
	})
	t.Run("Update", func(t *testing.T) {
		t.Parallel()

		id := hashStringToId("[TestCase] TestGetTestCase/Update")
		err := createAndWrite(path.Join(viper.GetString("path.test_cases"), strconv.Itoa(int(id)), "updated_at"), "")
		assert.NoError(t, err)
		err = createAndWrite(path.Join(viper.GetString("path.test_cases"), strconv.Itoa(int(id)), "in"), "test_get_test_case_update_input_old_content")
		assert.NoError(t, err)
		err = createAndWrite(path.Join(viper.GetString("path.test_cases"), strconv.Itoa(int(id)), "out"), "test_get_test_case_update_output_old_content")
		assert.NoError(t, err)

		oldStat, err := os.Stat(path.Join(viper.GetString("path.test_cases"), strconv.Itoa(int(id)), "updated_at"))
		assert.NoError(t, err)

		time.Sleep(time.Second) // ensure the file system record two different time for file updated_at

		task := api.Task{
			TestCaseID:        id,
			InputFile:         "fileURI/test_get_test_case_update_input/test_get_test_case_update_input_content",
			OutputFile:        "fileURI/test_get_test_case_update_output/test_get_test_case_update_output_content",
			TestCaseUpdatedAt: oldStat.ModTime().Add(time.Second),
		}
		task.GenerateFilePath()
		err = getTestCase(&task)
		assert.NoError(t, err)

		checkFile(t, path.Join(viper.GetString("path.test_cases"), strconv.Itoa(int(id)), "in"), "test_get_test_case_update_input_content")
		checkFile(t, path.Join(viper.GetString("path.test_cases"), strconv.Itoa(int(id)), "out"), "test_get_test_case_update_output_content")
		newStat, err := os.Stat(path.Join(viper.GetString("path.test_cases"), strconv.Itoa(int(id)), "updated_at"))
		assert.NoError(t, err)
		assert.True(t, oldStat.ModTime().Before(newStat.ModTime()))
	})
}

func TestBuild(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		task := api.Task{
			Language: models.Language{
				BuildScript: &models.Script{
					Name:      "test_build_success",
					UpdatedAt: time.Time{},
				},
			},
		}

		var err error
		task.JudgeDir, err = ioutil.TempDir("", "eduoj_judger_test_build_*")
		assert.NoError(t, err)
		err = createAndWrite(path.Join(task.JudgeDir, "code"), "test_build_success_code")
		assert.NoError(t, err)
		err = createAndWrite(path.Join(viper.GetString("path.scripts"), "test_build_success", "run"), `#!/bin/bash
echo [debug]test_build_success_output
echo code=$(cat $1/code) > $1/build_result
exit 0
`)
		assert.NoError(t, err)
		err = os.Chmod(path.Join(viper.GetString("path.scripts"), "test_build_success", "run"), 0755)
		assert.NoError(t, err)
		buildOutput, err := ioutil.TempFile("", "eduoj_judger_test_build_*")
		assert.NoError(t, err)
		task.BuildOutputPath = buildOutput.Name()
		assert.NoError(t, err)
		err = buildOutput.Close()
		assert.NoError(t, err)
		err = Build(&task)
		assert.NoError(t, err)
		checkFile(t, path.Join(task.JudgeDir, "build_result"), "code=test_build_success_code\n")
		checkFile(t, buildOutput.Name(), "[debug]test_build_success_output\n")
	})

	t.Run("Timeout", func(t *testing.T) {
		t.Parallel()
		task := api.Task{
			Language: models.Language{
				BuildScript: &models.Script{
					Name:      "test_build_timeout",
					UpdatedAt: time.Time{},
				},
			},
		}

		var err error
		task.JudgeDir, err = ioutil.TempDir("", "eduoj_judger_test_build_*")
		assert.NoError(t, err)
		err = createAndWrite(path.Join(task.JudgeDir, "code"), "test_build_timeout_code")
		assert.NoError(t, err)
		err = createAndWrite(path.Join(viper.GetString("path.scripts"), "test_build_timeout", "run"), `#!/bin/bash
echo [debug]test_build_timeout_output
echo code=$(cat $1/code) > $1/build_result
sleep 10m
exit 0
`)
		assert.NoError(t, err)
		err = os.Chmod(path.Join(viper.GetString("path.scripts"), "test_build_timeout", "run"), 0755)
		assert.NoError(t, err)
		buildOutput, err := ioutil.TempFile("", "eduoj_judger_test_build_*")
		assert.NoError(t, err)
		task.BuildOutputPath = buildOutput.Name()
		assert.NoError(t, err)
		err = buildOutput.Close()
		assert.NoError(t, err)
		err = Build(&task)
		assert.ErrorIs(t, err, ErrBuildError)
		checkFile(t, path.Join(task.JudgeDir, "build_result"), "code=test_build_timeout_code\n")
		checkFile(t, buildOutput.Name(), "[debug]test_build_timeout_output\n")
	})
	t.Run("NonZeroExitCode", func(t *testing.T) {
		t.Parallel()
		task := api.Task{
			Language: models.Language{
				BuildScript: &models.Script{
					Name:      "test_build_non_zero_exit_code",
					UpdatedAt: time.Time{},
				},
			},
		}

		var err error
		task.JudgeDir, err = ioutil.TempDir("", "eduoj_judger_test_build_*")
		assert.NoError(t, err)
		err = createAndWrite(path.Join(task.JudgeDir, "code"), "test_build_non_zero_exit_code_code")
		assert.NoError(t, err)
		err = createAndWrite(path.Join(viper.GetString("path.scripts"), "test_build_non_zero_exit_code", "run"), `#!/bin/bash
echo [debug]test_build_non_zero_exit_code_output
echo code=$(cat $1/code) > $1/build_result
exit 1
`)
		assert.NoError(t, err)
		err = os.Chmod(path.Join(viper.GetString("path.scripts"), "test_build_non_zero_exit_code", "run"), 0755)
		assert.NoError(t, err)
		buildOutput, err := ioutil.TempFile("", "eduoj_judger_test_build_*")
		assert.NoError(t, err)
		task.BuildOutputPath = buildOutput.Name()
		assert.NoError(t, err)
		err = buildOutput.Close()
		assert.NoError(t, err)
		err = Build(&task)
		assert.ErrorIs(t, err, ErrBuildError)
		checkFile(t, path.Join(task.JudgeDir, "build_result"), "code=test_build_non_zero_exit_code_code\n")
		checkFile(t, buildOutput.Name(), "[debug]test_build_non_zero_exit_code_output\n")
	})
	t.Run("OtherError", func(t *testing.T) {
		t.Parallel()
		task := api.Task{
			Language: models.Language{
				BuildScript: &models.Script{
					Name:      "test_build_other_error",
					UpdatedAt: time.Time{},
				},
			},
		}

		var err error
		task.JudgeDir, err = ioutil.TempDir("", "eduoj_judger_test_build_*")
		assert.NoError(t, err)
		err = createAndWrite(path.Join(task.JudgeDir, "code"), "test_build_other_error_code")
		assert.NoError(t, err)
		err = createAndWrite(path.Join(viper.GetString("path.scripts"), "test_build_other_error", "run"), `#!/bin/bash
exit 0
`)
		assert.NoError(t, err)
		buildOutput, err := ioutil.TempFile("", "eduoj_judger_test_build_*")
		assert.NoError(t, err)
		task.BuildOutputPath = buildOutput.Name()
		assert.NoError(t, err)
		err = buildOutput.Close()
		assert.NoError(t, err)
		err = Build(&task)
		assert.NotNil(t, err)
		assert.Regexp(t, regexp.MustCompile(`fail to build user program: fork/exec /tmp/eduoj_judger_test_scripts_\d+/test_build_other_error/run: permission denied`), err.Error())
	})
}

func TestRun(t *testing.T) { // TODO: fix race bug
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		task := api.Task{
			RunID:       hashStringToId("[Run] TestRun/Success"),
			TimeLimit:   1000,
			MemoryLimit: 500000000,
			Language: models.Language{
				Name: "cpp",
				RunScript: &models.Script{
					Name:      "test_run_success",
					UpdatedAt: time.Time{},
				},
			},
		}
		var err error
		task.JudgeDir, err = ioutil.TempDir("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		err = os.Chmod(path.Join(task.JudgeDir), 0777)
		assert.NoError(t, err)
		err = createAndWrite(path.Join(task.JudgeDir, "code.cpp"), `#include <iostream>
using namespace std;
int main(){
    char str[100];
    cin>>str;
    cout<<"test_run, input=";
    cout<<str<<endl;
    return 0;
}
`)
		assert.NoError(t, err)
		err = exec.Command("g++", path.Join(task.JudgeDir, "code.cpp"), "-o", path.Join(task.JudgeDir, "a.out")).Run()
		assert.NoError(t, err)
		err = createAndWrite(path.Join(viper.GetString("path.scripts"), "test_run_success", "run"), `#!/bin/bash
echo -n $1/a.out
`)
		assert.NoError(t, err)
		err = os.Chmod(path.Join(viper.GetString("path.scripts"), "test_run_success", "run"), 0777)
		assert.NoError(t, err)
		runFile, err := ioutil.TempFile("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		task.RunFilePath = runFile.Name()
		assert.NoError(t, err)
		err = runFile.Close()
		assert.NoError(t, err)
		inputFile, err := ioutil.TempFile("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		_, err = inputFile.WriteString("test_run_success_input")
		assert.NoError(t, err)
		task.InputFilePath = inputFile.Name()
		assert.NoError(t, err)
		err = inputFile.Close()
		assert.NoError(t, err)

		err = Run(&task)
		assert.NoError(t, err)
		checkFile(t, runFile.Name(), "test_run, input=test_run_success_input\n")
		checkFile(t, viper.GetString("log.sandbox_log_path"), "")
	})
	t.Run("TimeLimitExceeded", func(t *testing.T) {
		t.Parallel()
		task := api.Task{
			RunID:       hashStringToId("[Run] TestRun/TimeLimitExceeded"),
			TimeLimit:   1000,
			MemoryLimit: 10240000,
			Language: models.Language{
				Name: "cpp",
				RunScript: &models.Script{
					Name:      "test_run_time_limit_exceeded",
					UpdatedAt: time.Time{},
				},
			},
		}
		var err error
		task.JudgeDir, err = ioutil.TempDir("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		err = os.Chmod(path.Join(task.JudgeDir), 0777)
		assert.NoError(t, err)
		err = createAndWrite(path.Join(task.JudgeDir, "code.cpp"), `#include <iostream>
using namespace std;
int main(){
    char str[100];
    cin>>str;
    cout<<"test_run, input=";
    cout<<str<<endl;
	while(1);
    return 0;
}
`)
		assert.NoError(t, err)
		err = exec.Command("g++", path.Join(task.JudgeDir, "code.cpp"), "-o", path.Join(task.JudgeDir, "a.out")).Run()
		assert.NoError(t, err)
		err = createAndWrite(path.Join(viper.GetString("path.scripts"), "test_run_time_limit_exceeded", "run"), `#!/bin/bash
echo -n $1/a.out
`)
		assert.NoError(t, err)
		err = os.Chmod(path.Join(viper.GetString("path.scripts"), "test_run_time_limit_exceeded", "run"), 0777)
		assert.NoError(t, err)
		runFile, err := ioutil.TempFile("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		task.RunFilePath = runFile.Name()
		assert.NoError(t, err)
		err = runFile.Close()
		assert.NoError(t, err)
		inputFile, err := ioutil.TempFile("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		_, err = inputFile.WriteString("test_run_time_limit_exceeded_input")
		assert.NoError(t, err)
		task.InputFilePath = inputFile.Name()
		assert.NoError(t, err)
		err = inputFile.Close()
		assert.NoError(t, err)

		err = Run(&task)
		assert.ErrorIs(t, err, ErrTLE)
		checkFile(t, viper.GetString("log.sandbox_log_path"), "")
	})
	t.Run("MemoryLimitExceeded", func(t *testing.T) {
		t.Parallel()
		task := api.Task{
			RunID:       hashStringToId("[Run] TestRun/MemoryLimitExceeded"),
			TimeLimit:   1000,
			MemoryLimit: 102400,
			Language: models.Language{
				Name: "cpp",
				RunScript: &models.Script{
					Name:      "test_run_memory_limit_exceeded",
					UpdatedAt: time.Time{},
				},
			},
		}
		var err error
		task.JudgeDir, err = ioutil.TempDir("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		err = os.Chmod(path.Join(task.JudgeDir), 0777)
		assert.NoError(t, err)
		err = createAndWrite(path.Join(task.JudgeDir, "code.cpp"), `#include <iostream>
using namespace std;

int main(){
    char str[100];
    cin>>str;
    cout<<"test_run, input=";
    cout<<str<<endl;
    return 0;
}
`)
		assert.NoError(t, err)
		err = exec.Command("g++", path.Join(task.JudgeDir, "code.cpp"), "-o", path.Join(task.JudgeDir, "a.out")).Run()
		assert.NoError(t, err)
		err = createAndWrite(path.Join(viper.GetString("path.scripts"), "test_run_memory_limit_exceeded", "run"), `#!/bin/bash
echo -n $1/a.out
`)
		assert.NoError(t, err)
		err = os.Chmod(path.Join(viper.GetString("path.scripts"), "test_run_memory_limit_exceeded", "run"), 0777)
		assert.NoError(t, err)
		runFile, err := ioutil.TempFile("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		task.RunFilePath = runFile.Name()
		assert.NoError(t, err)
		err = runFile.Close()
		assert.NoError(t, err)
		inputFile, err := ioutil.TempFile("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		_, err = inputFile.WriteString("test_run_memory_limit_exceeded_input")
		assert.NoError(t, err)
		task.InputFilePath = inputFile.Name()
		assert.NoError(t, err)
		err = inputFile.Close()
		assert.NoError(t, err)

		err = Run(&task)
		assert.ErrorIs(t, err, ErrMLE)
		checkFile(t, viper.GetString("log.sandbox_log_path"), "")
	})
	t.Run("RuntimeError", func(t *testing.T) {
		t.Parallel()
		task := api.Task{
			RunID:       hashStringToId("[Run] TestRun/RuntimeError"),
			TimeLimit:   1000,
			MemoryLimit: 500000000,
			Language: models.Language{
				Name: "cpp",
				RunScript: &models.Script{
					Name:      "test_run_runtime_error",
					UpdatedAt: time.Time{},
				},
			},
		}
		var err error
		task.JudgeDir, err = ioutil.TempDir("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		err = os.Chmod(path.Join(task.JudgeDir), 0777)
		assert.NoError(t, err)
		err = createAndWrite(path.Join(task.JudgeDir, "code.cpp"), `#include <iostream>
using namespace std;

int main(){
    char str[100];
    cin>>str;
    cout<<"test_run, input=";
    cout<<str<<endl;
    int* p = nullptr;
    *p = 1;
    return 0;
}
`)
		assert.NoError(t, err)
		err = exec.Command("g++", path.Join(task.JudgeDir, "code.cpp"), "-o", path.Join(task.JudgeDir, "a.out")).Run()
		assert.NoError(t, err)
		err = createAndWrite(path.Join(viper.GetString("path.scripts"), "test_run_runtime_error", "run"), `#!/bin/bash
echo -n $1/a.out
`)
		assert.NoError(t, err)
		err = os.Chmod(path.Join(viper.GetString("path.scripts"), "test_run_runtime_error", "run"), 0777)
		assert.NoError(t, err)
		runFile, err := ioutil.TempFile("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		task.RunFilePath = runFile.Name()
		assert.NoError(t, err)
		err = runFile.Close()
		assert.NoError(t, err)
		inputFile, err := ioutil.TempFile("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		_, err = inputFile.WriteString("test_run_runtime_error_input")
		assert.NoError(t, err)
		task.InputFilePath = inputFile.Name()
		assert.NoError(t, err)
		err = inputFile.Close()
		assert.NoError(t, err)

		err = Run(&task)
		assert.ErrorIs(t, err, ErrRTE)
		checkFile(t, viper.GetString("log.sandbox_log_path"), "")
	})
	t.Run("DangerousSystemCalls", func(t *testing.T) {
		t.Parallel()
		task := api.Task{
			RunID:       hashStringToId("[Run] TestRun/DangerousSystemCalls"),
			TimeLimit:   1000,
			MemoryLimit: 10240000,
			Language: models.Language{
				Name: "cpp",
				RunScript: &models.Script{
					Name:      "test_run_dangerous_system_calls",
					UpdatedAt: time.Time{},
				},
			},
		}
		var err error
		task.JudgeDir, err = ioutil.TempDir("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		err = createAndWrite(path.Join(task.JudgeDir, "code.cpp"), `#include <iostream>
#include <cstdlib>
using namespace std;

int main(){
    char str[100];
    cin>>str;
    cout<<"test_run, input=";
    cout<<str<<endl;
    system("ls");
    return 0;
}
`)
		assert.NoError(t, err)
		err = exec.Command("g++", path.Join(task.JudgeDir, "code.cpp"), "-o", path.Join(task.JudgeDir, "a.out")).Run()
		assert.NoError(t, err)
		err = createAndWrite(path.Join(viper.GetString("path.scripts"), "test_run_dangerous_system_calls", "run"), `#!/bin/bash
echo -n $1/a.out
`)
		assert.NoError(t, err)
		err = os.Chmod(path.Join(viper.GetString("path.scripts"), "test_run_dangerous_system_calls", "run"), 0777)
		assert.NoError(t, err)
		runFile, err := ioutil.TempFile("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		task.RunFilePath = runFile.Name()
		assert.NoError(t, err)
		err = runFile.Close()
		assert.NoError(t, err)
		inputFile, err := ioutil.TempFile("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		_, err = inputFile.WriteString("test_run_dangerous_system_calls_input")
		assert.NoError(t, err)
		task.InputFilePath = inputFile.Name()
		assert.NoError(t, err)
		err = inputFile.Close()
		assert.NoError(t, err)
		err = os.Chmod(path.Join(task.JudgeDir), 0777)
		assert.NoError(t, err)

		err = Run(&task)
		assert.ErrorIs(t, err, ErrDSC)
		checkFile(t, viper.GetString("log.sandbox_log_path"), "")
	})
	t.Run("SystemError", func(t *testing.T) {
		// Not Parallel
		logPath := viper.GetString("log.sandbox_log_path")
		logFile, err := ioutil.TempFile("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		viper.Set("log.sandbox_log_path", logFile.Name())
		err = logFile.Close()
		assert.NoError(t, err)
		t.Cleanup(func() {
			viper.Set("log.sandbox_log_path", logPath)
		})

		task := api.Task{
			RunID:       hashStringToId("[Run] TestRun/SystemError"),
			TimeLimit:   1000,
			MemoryLimit: 500000000,
			Language: models.Language{
				Name: "cpp",
				RunScript: &models.Script{
					Name:      "test_run_system_error",
					UpdatedAt: time.Time{},
				},
			},
		}
		task.JudgeDir, err = ioutil.TempDir("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		err = os.Chmod(path.Join(task.JudgeDir), 0777)
		assert.NoError(t, err)
		err = createAndWrite(path.Join(task.JudgeDir, "code.cpp"), `#include <iostream>
#include <cstdlib>
using namespace std;

int main(){
  char str[100];
  cin>>str;
  cout<<"test_run, input=";
  cout<<str<<endl;
  return 0;
}
`)
		assert.NoError(t, err)
		err = exec.Command("g++", path.Join(task.JudgeDir, "code.cpp"), "-o", path.Join(task.JudgeDir, "a.out")).Run()
		assert.NoError(t, err)
		err = createAndWrite(path.Join(viper.GetString("path.scripts"), "test_run_system_error", "run"), `#!/bin/bash
echo -n $1/a.out
`)
		assert.NoError(t, err)
		err = os.Chmod(path.Join(viper.GetString("path.scripts"), "test_run_system_error", "run"), 0777)
		assert.NoError(t, err)
		runFile, err := ioutil.TempFile("", "eduoj_judger_test_run_*")
		assert.NoError(t, err)
		task.RunFilePath = runFile.Name()
		assert.NoError(t, err)
		err = runFile.Close()
		assert.NoError(t, err)
		err = Run(&task)
		assert.NotNil(t, err)
		assert.Equal(t, "runtime error", err.Error())
		l, err := os.Open(viper.GetString("log.sandbox_log_path"))
		assert.NoError(t, err)
		b, err := ioutil.ReadAll(l)
		assert.NoError(t, err)
		assert.Regexp(t, regexp.MustCompile(`FATAL \[\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d] \[child\.c:87]Error: System errno: No such file or directory; Internal errno: DUP2_FAILED`), string(b))
		err = l.Close()
		assert.NoError(t, err)
	})
}

func TestHashOutput(t *testing.T) {
	t.Parallel()

	runFile, err := ioutil.TempFile("", "eduoj_judger_test_hash_output_*")
	assert.NoError(t, err)
	_, err = runFile.WriteString("tes    t_h as   h_run\n\n _ f  ile_c   on t  e n t   \n \n")
	assert.NoError(t, err)
	err = runFile.Close()
	assert.NoError(t, err)

	task := api.Task{
		RunFilePath: runFile.Name(),
	}
	err = hashOutput(&task)
	assert.NoError(t, err)

	h := sha256.Sum256([]byte("test_hash_run_file_content"))
	assert.Equal(t, hex.EncodeToString(h[:]), task.OutputStrippedHash)
}

func TestCompare(t *testing.T) {
	err := os.MkdirAll(path.Join(viper.GetString("path.scripts"), "test_compare_script"), 0700)
	assert.NoError(t, err)
	r, err := os.Create(path.Join(viper.GetString("path.scripts"), "test_compare_script", "run"))
	assert.NoError(t, err)
	err = os.Chmod(r.Name(), 0700)
	assert.NoError(t, err)
	_, err = r.WriteString(`#!/bin/bash
#echo 1
#echo $1
#echo $2
#echo $(cat $1)
#echo $(cat $2)

ret=$(diff -w $1 $2)
# echo ==[$ret]==
content1=$(cat $1)
if [ "$content1" == "OTHER_OUTPUT" ]
then
  exit 3
elif [ "$ret" != "" ]
then
  exit 1
fi
critical=$(diff $1 $2)
# echo ==[$critical]==
if  [ "$critical" == "" ]
then
  exit 0
else
  exit 2
fi
`)
	assert.NoError(t, err)
	err = r.Close()
	assert.NoError(t, err)

	t.Run("Same", func(t *testing.T) {
		t.Parallel()
		runFile, err := ioutil.TempFile("", "eduoj_judger_test_compare_*")
		assert.NoError(t, err)
		_, err = runFile.WriteString("test_compare_same")
		assert.NoError(t, err)
		err = runFile.Close()
		assert.NoError(t, err)

		err = createAndWrite(path.Join(viper.GetString("path.test_cases"), "test_compare_script_same", "out"), "test_compare_same")
		assert.NoError(t, err)

		compareOutputFile, err := ioutil.TempFile("", "eduoj_judger_test_compare_*")
		assert.NoError(t, err)
		err = compareOutputFile.Close()
		assert.NoError(t, err)

		task := api.Task{
			OutputFilePath: path.Join(viper.GetString("path.test_cases"), "test_compare_script_same", "out"),
			RunFilePath:    runFile.Name(),
			CompareScript: models.Script{
				Name:      "test_compare_script",
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			CompareOutputPath: compareOutputFile.Name(),
		}
		err = Compare(&task)
		assert.NoError(t, err)
	})
	t.Run("Different", func(t *testing.T) {
		t.Parallel()
		runFile, err := ioutil.TempFile("", "eduoj_judger_test_compare_*")
		assert.NoError(t, err)
		_, err = runFile.WriteString("test_compare_run")
		assert.NoError(t, err)
		err = runFile.Close()
		assert.NoError(t, err)

		err = createAndWrite(path.Join(viper.GetString("path.test_cases"), "test_compare_script_different", "out"), "test_compare_output")
		assert.NoError(t, err)

		compareOutputFile, err := ioutil.TempFile("", "eduoj_judger_test_compare_*")
		assert.NoError(t, err)
		err = compareOutputFile.Close()
		assert.NoError(t, err)

		task := api.Task{
			OutputFilePath: path.Join(viper.GetString("path.test_cases"), "test_compare_script_different", "out"),
			RunFilePath:    runFile.Name(),
			CompareScript: models.Script{
				Name:      "test_compare_script",
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			CompareOutputPath: compareOutputFile.Name(),
		}
		err = Compare(&task)
		assert.Equal(t, ErrWA, err)
	})
	t.Run("OtherOutput", func(t *testing.T) {
		t.Parallel()
		runFile, err := ioutil.TempFile("", "eduoj_judger_test_compare_*")
		assert.NoError(t, err)
		_, err = runFile.WriteString("OTHER_OUTPUT")
		assert.NoError(t, err)
		err = runFile.Close()
		assert.NoError(t, err)

		err = createAndWrite(path.Join(viper.GetString("path.test_cases"), "test_compare_script_other_output", "out"), "test_compare_other_output")
		assert.NoError(t, err)

		compareOutputFile, err := ioutil.TempFile("", "eduoj_judger_test_compare_*")
		assert.NoError(t, err)
		err = compareOutputFile.Close()
		assert.NoError(t, err)

		task := api.Task{
			OutputFilePath: path.Join(viper.GetString("path.test_cases"), "test_compare_script_other_output", "out"),
			RunFilePath:    runFile.Name(),
			CompareScript: models.Script{
				Name:      "test_compare_script",
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			CompareOutputPath: compareOutputFile.Name(),
		}
		err = Compare(&task)
		assert.NotNil(t, err)
		assert.Equal(t, "unexpected compare script output: 3", err.Error())
	})
	t.Run("PresentationError", func(t *testing.T) {
		t.Parallel()
		runFile, err := ioutil.TempFile("", "eduoj_judger_test_compare_*")
		assert.NoError(t, err)
		_, err = runFile.WriteString("test_  compare_  run")
		assert.NoError(t, err)
		err = runFile.Close()
		assert.NoError(t, err)

		err = createAndWrite(path.Join(viper.GetString("path.test_cases"), "test_compare_script_presentation_error", "out"), "t e     st  _co       mp a r e_r un   ")
		assert.NoError(t, err)

		compareOutputFile, err := ioutil.TempFile("", "eduoj_judger_test_compare_*")
		assert.NoError(t, err)
		err = compareOutputFile.Close()
		assert.NoError(t, err)

		task := api.Task{
			OutputFilePath: path.Join(viper.GetString("path.test_cases"), "test_compare_script_presentation_error", "out"),
			RunFilePath:    runFile.Name(),
			CompareScript: models.Script{
				Name:      "test_compare_script",
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
			CompareOutputPath: compareOutputFile.Name(),
		}
		err = Compare(&task)
		assert.Equal(t, ErrPE, err)
	})
}
