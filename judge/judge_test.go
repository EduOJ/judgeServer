package judge

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func checkFile(t *testing.T, path, content string) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0666)
	assert.NoError(t, err)
	b := make([]byte, len(content))
	_, err = file.Read(b)
	assert.NoError(t, err)
	assert.Equal(t, content, string(b))
}

func checkFileNonExist(t *testing.T, path string) {
	_, err := os.OpenFile(path, os.O_RDONLY, 0666)
	assert.True(t, os.IsNotExist(err))
}

func TestMain(m *testing.M) {
	config := `user:
  compile: build_user
  run: run_user
path:
  scripts: ../test_file/scripts
  test_cases: ../test_file/test_cases
  temp: ../test_file/temp
`
	viper.SetConfigType("yml")
	if err := viper.ReadConfig(strings.NewReader(config)); err != nil {
		panic(err)
	}
	dir, err := ioutil.TempDir("", "eduoj_judger_test_scripts_*")
	if err != nil {
		panic(errors.Wrap(err, "could not create temp dir"))
	}
	viper.Set("path.scripts", dir)
	dir, err = ioutil.TempDir("", "eduoj_judger_test_test_cases_*")
	if err != nil {
		panic(errors.Wrap(err, "could not create temp dir"))
	}
	viper.Set("path.test_cases", dir)
	ret := m.Run()
	os.Exit(ret)
}
