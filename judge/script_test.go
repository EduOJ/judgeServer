package judge_test

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/suntt2019/EduOJJudger/judge"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func createFileForTest(path string, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0777); err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write([]byte(content))
	return err
}

func createZipFileForTest(target string, files ...string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0777); err != nil {
		return err
	}
	return exec.Command("zip", append([]string{"-j", target}, files...)...).Run()
}

func TestInstallScript(t *testing.T) {
	t.Parallel()
	scriptsDir := viper.GetString("path.scripts")
	assert.Nil(t, createFileForTest(scriptsDir+"/compile", `echo "test_install_script_content" > t.txt`+"\n"))
	assert.Nil(t, createFileForTest(scriptsDir+"/other_file", `other file for testing install script content`+"\n"))

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, createZipFileForTest(scriptsDir+"/downloads/test_install_script_success.zip", scriptsDir+"/compile", scriptsDir+"/other_file"))
		assert.Nil(t, createFileForTest(scriptsDir+"/test_install_script_success/old.txt", "test_install_script_success_old_content"))
		info, err := os.Stat(scriptsDir + "/test_install_script_success")
		assert.Nil(t, err)
		assert.Nil(t, judge.InstallScript("test_install_script_success", info.ModTime().Add(time.Second)))
		checkFile(t, scriptsDir+"/test_install_script_success/compile", `echo "test_install_script_content" > t.txt`+"\n")
		checkFile(t, scriptsDir+"/test_install_script_success/other_file", `other file for testing install script content`+"\n")
		checkFile(t, scriptsDir+"/test_install_script_success/t.txt", "test_install_script_content")
		checkFileNonExist(t, scriptsDir+"/test_install_script_success/old.txt")
	})

	t.Run("Expired", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, createZipFileForTest(scriptsDir+"/downloads/test_install_script_expired.zip", scriptsDir+"/compile", scriptsDir+"/other_file"))
		assert.Nil(t, createFileForTest(scriptsDir+"/test_install_script_expired/old.txt", "test_install_script_expired_old_content"))
		info, err := os.Stat(scriptsDir + "/test_install_script_expired")
		assert.Nil(t, err)
		assert.Equal(t, "already up to date", judge.InstallScript("test_install_script_expired", info.ModTime().Add(-1*time.Second)).Error())
		checkFileNonExist(t, scriptsDir+"/test_install_script_expired/compile")
		checkFileNonExist(t, scriptsDir+"/test_install_script_expired/other_file")
		checkFileNonExist(t, scriptsDir+"/test_install_script_expired/t.txt")
		checkFile(t, scriptsDir+"/test_install_script_expired/old.txt", "test_install_script_expired_old_content")
	})
}

func TestRunScript(t *testing.T) {
	t.Parallel()
	scriptsDir := viper.GetString("path.scripts")
	assert.Nil(t, createFileForTest(scriptsDir+"/test_run_script/run", "a=`cat $1`\nb=`cat $2`\necho $a$b > out.txt\n"))
	assert.Nil(t, createFileForTest("../test_file/test_run_script/a.txt", "str1"))
	assert.Nil(t, createFileForTest("../test_file/test_run_script/b.txt", "str2"))
	err := judge.RunScript("test_run_script", "../test_file/test_run_script", "a.txt", "b.txt")
	assert.Nil(t, err)
	checkFile(t, "../test_file/test_run_script/out.txt", "str1str2")
}
