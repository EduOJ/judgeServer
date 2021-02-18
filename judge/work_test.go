package judge

import (
	"github.com/leoleoasd/EduOJBackend/database/models"
	"github.com/minio/sha256-simd"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/suntt2019/EduOJJudger/api"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"
	"time"
)

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
	assert.Equal(t, string(h[:]), task.OutputStrippedHash)
}

func TestCompare(t *testing.T) {
	err := os.MkdirAll(path.Join(viper.GetString("path.scripts"), "test_compare_script"), 0777)
	assert.NoError(t, err)
	r, err := os.Create(path.Join(viper.GetString("path.scripts"), "test_compare_script", "run"))
	assert.NoError(t, err)
	err = os.Chmod(r.Name(), 0777)
	assert.NoError(t, err)
	_, err = r.WriteString(`#!/bin/bash
ret=$(diff $1 $2)
content1=$(cat $1)
#content2=$(cat $2)
#echo 0:$0
#echo 1:$1
#echo 2:$2
#echo c1:$content1
#echo c2:$content2

if [ "$content1" == "OTHER_OUTPUT" ]
then
  echo "OTHER_OUTPUT"
elif [ "$ret" == "" ]
then
  echo "ACCEPTED"
else
  echo "WRONG_ANSWER"
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

		task := api.Task{
			OutputFilePath: path.Join(viper.GetString("path.test_cases"), "test_compare_script_same", "out"),
			RunFilePath:    runFile.Name(),
			CompareScript: models.Script{
				Name:      "test_compare_script",
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
		}
		same, err := compare(&task)
		assert.NoError(t, err)
		assert.True(t, same)
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

		task := api.Task{
			OutputFilePath: path.Join(viper.GetString("path.test_cases"), "test_compare_script_different", "out"),
			RunFilePath:    runFile.Name(),
			CompareScript: models.Script{
				Name:      "test_compare_script",
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
		}
		same, err := compare(&task)
		assert.NoError(t, err)
		assert.False(t, same)
	})

	t.Run("OtherOutput", func(t *testing.T) {
		t.Parallel()
		runFile, err := ioutil.TempFile("", "eduoj_judger_test_compare_*")
		assert.NoError(t, err)
		_, err = runFile.WriteString("test_compare_other_output_run")
		assert.NoError(t, err)
		err = runFile.Close()
		assert.NoError(t, err)

		err = createAndWrite(path.Join(viper.GetString("path.test_cases"), "test_compare_script_other_output", "out"), "OTHER_OUTPUT")
		assert.NoError(t, err)

		task := api.Task{
			OutputFilePath: path.Join(viper.GetString("path.test_cases"), "test_compare_script_other_output", "out"),
			RunFilePath:    runFile.Name(),
			CompareScript: models.Script{
				Name:      "test_compare_script",
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
		}
		same, err := compare(&task)
		assert.NoError(t, err)
		assert.False(t, same)
	})
}
