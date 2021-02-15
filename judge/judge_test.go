package judge_test

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/suntt2019/EduOJJudger/base"
	"os"
	"strings"
	"testing"
)

func checkFile(t *testing.T, path, content string) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	assert.Nil(t, err)
	b := make([]byte, len(content))
	_, err = file.Read(b)
	assert.Nil(t, err)
	assert.Equal(t, content, string(b))
}

func checkFileNonExist(t *testing.T, path string) {
	_, err := os.OpenFile(path, os.O_RDONLY, 0666)
	assert.True(t, os.IsNotExist(err))
}

func TestMain(m *testing.M) {
	config := `user:
  script: script_user
  compile: compile_user
  run: run_user
  compare: compare_user
path:
  scripts: ../test_file/scripts
  runs: ../test_file/runs
  temp: ../test_file/temp
timeout:
  script:
    unzip: 300s
    compile: 600s
`
	viper.SetConfigType("yml")
	if err := viper.ReadConfig(strings.NewReader(config)); err != nil {
		panic(err)
	}
	if err := base.ScriptUser.Init(viper.GetString("user.script")); err != nil {
		panic(err)
	}
	ret := m.Run()
	os.Exit(ret)
}
