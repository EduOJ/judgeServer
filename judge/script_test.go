package judge

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"
)

func TestCheckScript(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, os.Mkdir(path.Join(viper.GetString("path.scripts"), "test_check_script_success"), 0777))
		stat, err := os.Stat(path.Join(viper.GetString("path.scripts"), "test_check_script_success"))
		assert.NoError(t, err)
		ok, err := checkScript("test_check_script_success", stat.ModTime().Add(-1*time.Second))
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("Expired", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, os.Mkdir(path.Join(viper.GetString("path.scripts"), "test_check_script_expired"), 0777))
		stat, err := os.Stat(path.Join(viper.GetString("path.scripts"), "test_check_script_expired"))
		assert.NoError(t, err)
		ok, err := checkScript("test_check_script_success", stat.ModTime().Add(time.Second))
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}

func TestInstallScript(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		tempDir, err := ioutil.TempDir("", "eduoj_judger_test_install_script_*")
		assert.NoError(t, err)

		compile, err := os.Create(path.Join(tempDir, "compile"))
		assert.NoError(t, err)
		_, err = compile.WriteString(`#!/bin/bash
echo "test_install_script_success" > t.txt`)
		assert.NoError(t, err)
		err = os.Chmod(compile.Name(), 0777)
		assert.NoError(t, err)
		err = compile.Close()
		assert.NoError(t, err)

		other, err := os.Create(path.Join(tempDir, "other_file"))
		assert.NoError(t, err)
		_, err = other.WriteString("test_install_script_success_other")
		assert.NoError(t, err)
		err = other.Close()
		assert.NoError(t, err)

		tempFile, err := ioutil.TempFile("", "eduoj_judger_test_install_script_*")
		assert.NoError(t, err)
		err = exec.Command("zip", "-j", tempFile.Name(), compile.Name(), other.Name()).Run()
		assert.NoError(t, err)

		err = os.MkdirAll(path.Join(viper.GetString("path.scripts"), "test_install_script_success"), 0777)
		assert.NoError(t, err)
		old, err := os.Create(path.Join(viper.GetString("path.scripts"), "test_install_script_success", "old_file"))
		assert.NoError(t, err)
		_, err = old.WriteString("test_install_script_success_old_file")
		assert.NoError(t, err)
		err = old.Close()
		assert.NoError(t, err)

		err = installScript("test_install_script_success", tempFile)
		assert.NoError(t, err)
		checkFile(t, path.Join(viper.GetString("path.scripts"), "test_install_script_success", "compile"),
			`#!/bin/bash
echo "test_install_script_success" > t.txt`)
		checkFile(t, path.Join(viper.GetString("path.scripts"), "test_install_script_success", "other_file"),
			"test_install_script_success_other")
		checkFile(t, path.Join(viper.GetString("path.scripts"), "test_install_script_success", "t.txt"),
			"test_install_script_success")
		checkFileNonExist(t, path.Join(viper.GetString("path.scripts"), "test_install_script_success", "old_file"))
	})

	t.Run("CompileCouldNotRun", func(t *testing.T) {
		t.Parallel()

		tempDir, err := ioutil.TempDir("", "eduoj_judger_test_install_script_*")
		assert.NoError(t, err)

		compile, err := os.Create(path.Join(tempDir, "compile"))
		assert.NoError(t, err)
		_, err = compile.WriteString("command_that_could_not_run")
		assert.NoError(t, err)
		err = os.Chmod(compile.Name(), 0777)
		assert.NoError(t, err)
		err = compile.Close()
		assert.NoError(t, err)

		other, err := os.Create(path.Join(tempDir, "other_file"))
		assert.NoError(t, err)
		_, err = other.WriteString("test_install_script_compile_could_not_run_other")
		assert.NoError(t, err)
		err = other.Close()
		assert.NoError(t, err)

		tempFile, err := ioutil.TempFile("", "eduoj_judger_test_install_script_*")
		assert.NoError(t, err)
		err = exec.Command("zip", "-j", tempFile.Name(), compile.Name(), other.Name()).Run()
		assert.NoError(t, err)

		err = installScript("test_install_script_compile_could_not_run", tempFile)
		assert.NotNil(t, err)
		assert.Equal(t, "could not compile script: fork/exec ./compile: exec format error", err.Error())
	})
}

// tests for function judge.runScript
func TestRunScript(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()

		err := os.MkdirAll(path.Join(viper.GetString("path.scripts"), "test_run_script_success"), 0777)
		assert.NoError(t, err)
		r, err := os.Create(path.Join(viper.GetString("path.scripts"), "test_run_script_success", "run"))
		assert.NoError(t, err)
		err = os.Chmod(r.Name(), 0777)
		assert.NoError(t, err)
		_, err = r.WriteString(`#!/bin/bash
echo "test_run_script_success" > t.txt`)
		assert.NoError(t, err)
		err = r.Close()
		assert.NoError(t, err)

		err = runScript("test_run_script_success")
		assert.NoError(t, err)
		checkFile(t, path.Join(viper.GetString("path.scripts"), "test_run_script_success", "t.txt"),
			"test_run_script_success")
	})

	t.Run("Fail", func(t *testing.T) {
		t.Parallel()

		err := os.MkdirAll(path.Join(viper.GetString("path.scripts"), "test_run_script_fail"), 0777)
		assert.NoError(t, err)
		r, err := os.Create(path.Join(viper.GetString("path.scripts"), "test_run_script_fail", "run"))
		assert.NoError(t, err)
		err = os.Chmod(r.Name(), 0777)
		assert.NoError(t, err)
		_, err = r.WriteString("command_that_could_not_run")
		assert.NoError(t, err)
		err = r.Close()
		assert.NoError(t, err)

		err = runScript("test_run_script_fail")
		assert.NotNil(t, err)
		assert.Equal(t, "could not run script: fork/exec ./run: exec format error", err.Error())
	})
}
