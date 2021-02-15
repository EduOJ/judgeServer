package judge_test

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/suntt2019/EduOJJudger/base"
	"github.com/suntt2019/EduOJJudger/judge"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
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

func TestCheckScript(t *testing.T) {
	t.Parallel()
	scriptsDir := viper.GetString("path.scripts")
	t.Run("UpToDate", func(t *testing.T) {
		assert.NoError(t, os.RemoveAll(path.Join(scriptsDir, "test_check_script_up_to_date")))
		t.Cleanup(func() {
			assert.NoError(t, os.RemoveAll(path.Join(scriptsDir, "test_check_script_up_to_date")))
		})
		assert.NoError(t, createFileForTest(path.Join(scriptsDir, "test_check_script_up_to_date/t.txt"), "test_check_script_up_to_date_content"))
		status, err := os.Stat(path.Join(scriptsDir, "test_check_script_up_to_date"))
		assert.NoError(t, err)
		ok, err := judge.CheckScript("test_check_script_up_to_date", status.ModTime().Add(-1*time.Second))
		assert.True(t, ok)
		assert.NoError(t, err)
	})
	t.Run("Expired", func(t *testing.T) {
		assert.NoError(t, os.RemoveAll(path.Join(scriptsDir, "test_check_script_expired")))
		t.Cleanup(func() {
			assert.NoError(t, os.RemoveAll(path.Join(scriptsDir, "test_check_script_expired")))
		})
		assert.NoError(t, createFileForTest(path.Join(scriptsDir, "test_check_script_expired/t.txt"), "test_check_script_expired_content"))
		status, err := os.Stat(path.Join(scriptsDir, "test_check_script_expired"))
		assert.NoError(t, err)
		ok, err := judge.CheckScript("test_check_script_expired", status.ModTime().Add(1*time.Second))
		assert.False(t, ok)
		assert.NoError(t, err)
	})
}

func TestInstallScript(t *testing.T) {
	t.Parallel()
	scriptsDir := viper.GetString("path.scripts")
	TestFileTempDir, err := ioutil.TempDir(viper.GetString("path.temp"), "")
	assert.NoError(t, err)
	assert.NoError(t, os.RemoveAll(path.Join(scriptsDir, "test_install_script_success")))
	t.Cleanup(func() {
		assert.NoError(t, os.RemoveAll(path.Join(scriptsDir, "test_install_script_success")))
	})

	assert.NoError(t, createFileForTest(path.Join(TestFileTempDir, "compile"), `#!/bin/bash
echo "test_install_script_content" > t.txt
`))
	assert.NoError(t, createFileForTest(path.Join(TestFileTempDir, "other_file"), "other file for testing install script content\n"))
	dir, err := ioutil.TempDir(viper.GetString("path.temp"), "")
	assert.NoError(t, err)
	assert.NoError(t, createZipFileForTest(path.Join(dir, "test_install_script_success.zip"), path.Join(TestFileTempDir, "compile"), path.Join(TestFileTempDir, "other_file")))
	assert.NoError(t, base.ScriptUser.OwnRWX(dir))
	assert.NoError(t, base.ScriptUser.OwnRWX(dir+"/test_install_script_success.zip"))
	assert.NoError(t, createFileForTest(path.Join(scriptsDir, "test_install_script_success/old.txt"), "test_install_script_success_old_content"))
	assert.NoError(t, judge.InstallScript("test_install_script_success", dir))
	checkFile(t, path.Join(scriptsDir, "test_install_script_success/compile"), `#!/bin/bash
echo "test_install_script_content" > t.txt
`)
	checkFile(t, path.Join(scriptsDir, "test_install_script_success/other_file"), "other file for testing install script content\n")
	checkFile(t, path.Join(scriptsDir, "test_install_script_success/t.txt"), "test_install_script_content")
	checkFileNonExist(t, path.Join(scriptsDir, "test_install_script_success/old.txt"))
}

func TestRunScript(t *testing.T) {
	t.Parallel()
	scriptsDir := viper.GetString("path.scripts")
	TestFileTempDir, err := ioutil.TempDir(viper.GetString("path.temp"), "")
	assert.NoError(t, err)
	assert.NoError(t, os.RemoveAll(path.Join(scriptsDir, "test_run_script")))
	t.Cleanup(func() {
		assert.NoError(t, os.RemoveAll(path.Join(scriptsDir, "test_run_script")))
	})
	assert.NoError(t, createFileForTest(path.Join(scriptsDir, "test_run_script/run"), "#!/bin/bash\na=`cat $1`\nb=`cat $2`\necho $a$b > out.txt\n"))
	assert.NoError(t, base.ScriptUser.OwnRWX(path.Join(scriptsDir, "test_run_script/run")))
	assert.NoError(t, createFileForTest(path.Join(TestFileTempDir, "a.txt"), "str1"))
	assert.NoError(t, createFileForTest(path.Join(TestFileTempDir, "b.txt"), "str2"))
	assert.NoError(t, base.ScriptUser.OwnRWX(path.Join(TestFileTempDir)))
	err = judge.RunScript("test_run_script", TestFileTempDir, "a.txt", "b.txt")
	assert.NoError(t, err)
	checkFile(t, path.Join(TestFileTempDir, "out.txt"), "str1str2")
}
