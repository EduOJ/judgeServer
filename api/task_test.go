package api

import (
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/backend/database/models"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"path"
	"testing"
	"time"
)

func TestGenerateFilePath(t *testing.T) {
	task := Task{
		TestCaseID:     0,
		InputFilePath:  "",
		OutputFilePath: "",
	}
	task.GenerateFilePath()
	assert.Equal(t, path.Join(viper.GetString("path.test_cases"), "0/in"), task.InputFilePath)
	assert.Equal(t, path.Join(viper.GetString("path.test_cases"), "0/out"), task.OutputFilePath)
}

var taskStatus string

func task(wr http.ResponseWriter, q url.Values) {
	if q.Get("poll") != "1" {
		panic("poll not equals to 1")
	}
	switch taskStatus {
	case "Ready":
		resp := response.GetTaskResponse{
			Message: "SUCCESS",
			Error:   nil,
			Data: struct {
				RunID             uint            `json:"run_id"`
				Language          models.Language `json:"language"`
				TestCaseID        uint            `json:"test_case_id"`
				InputFile         string          `json:"input_file"`  // pre-signed url
				OutputFile        string          `json:"output_file"` // same as above
				CodeFile          string          `json:"code_file"`
				TestCaseUpdatedAt time.Time       `json:"test_case_updated_at"`
				MemoryLimit       uint64          `json:"memory_limit"` // Byte
				TimeLimit         uint            `json:"time_limit"`   // ms
				BuildArg          string          `json:"build_arg"`    // E.g.  O2=false
				CompareScript     models.Script   `json:"compare_script"`
			}{
				RunID: 0,
				Language: models.Language{
					Name:             "test_task_language",
					ExtensionAllowed: nil,
					BuildScriptName:  "test_task_build_script",
					BuildScript: &models.Script{
						Name:      "test_task_build_script",
						Filename:  "test_task_build_script_filename",
						CreatedAt: hashStringToTime("test_task_build_script_created_at"),
						UpdatedAt: hashStringToTime("test_task_build_script_updated_at"),
					},
					RunScriptName: "test_task_run_script",
					RunScript: &models.Script{
						Name:      "test_task_run_script",
						Filename:  "test_task_run_script_filename",
						CreatedAt: hashStringToTime("test_task_run_script_created_at"),
						UpdatedAt: hashStringToTime("test_task_run_script_updated_at"),
					},
					CreatedAt: hashStringToTime("test_task_language_created_at"),
					UpdatedAt: hashStringToTime("test_task_language_updated_at"),
				},
				TestCaseID:        1,
				InputFile:         "test_task_input_file_presigned_url",
				OutputFile:        "test_task_output_file_presigned_url",
				CodeFile:          "test_task_code_file_presigned_url",
				TestCaseUpdatedAt: hashStringToTime("test_task_test_case_updated_at"),
				MemoryLimit:       1024,
				TimeLimit:         1000,
				BuildArg:          "test,task,build,args",
				CompareScript: models.Script{
					Name:      "test_task_compare_script",
					Filename:  "test_task_compare_script_filename",
					CreatedAt: hashStringToTime("test_task_compare_script_created_at"),
					UpdatedAt: hashStringToTime("test_task_compare_script_updated_at"),
				},
			},
		}
		if err := marshalAndWrite(wr, resp); err != nil {
			panic(err)
		}
	case "NotFound":
		wr.WriteHeader(http.StatusNotFound)
		resp := response.Response{
			Message: "NOT_FOUND",
			Error:   nil,
			Data:    nil,
		}
		if err := marshalAndWrite(wr, resp); err != nil {
			panic(err)
		}
	default:
		panic("unexpected task status")
	}
}

func TestGetTask(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		// Not Parallel
		taskStatus = "Ready"
		task, err := GetTask()
		assert.NoError(t, err)
		assert.Equal(t, &Task{
			RunID: 0,
			Language: models.Language{
				Name:             "test_task_language",
				ExtensionAllowed: nil,
				BuildScriptName:  "",
				BuildScript: &models.Script{
					Name:      "test_task_build_script",
					Filename:  "test_task_build_script_filename",
					CreatedAt: hashStringToTime("test_task_build_script_created_at"),
					UpdatedAt: hashStringToTime("test_task_build_script_updated_at"),
				},
				RunScriptName: "",
				RunScript: &models.Script{
					Name:      "test_task_run_script",
					Filename:  "test_task_run_script_filename",
					CreatedAt: hashStringToTime("test_task_run_script_created_at"),
					UpdatedAt: hashStringToTime("test_task_run_script_updated_at"),
				},
				CreatedAt: hashStringToTime("test_task_language_created_at"),
				UpdatedAt: hashStringToTime("test_task_language_updated_at"),
			},
			TestCaseID:        1,
			InputFile:         "test_task_input_file_presigned_url",
			OutputFile:        "test_task_output_file_presigned_url",
			InputFilePath:     path.Join(viper.GetString("path.test_cases"), "1", "in"),
			OutputFilePath:    path.Join(viper.GetString("path.test_cases"), "1", "out"),
			CodeFile:          "test_task_code_file_presigned_url",
			TestCaseUpdatedAt: hashStringToTime("test_task_test_case_updated_at"),
			MemoryLimit:       1024,
			TimeLimit:         1000,
			BuildArg:          "test,task,build,args",
			CompareScript: models.Script{
				Name:      "test_task_compare_script",
				Filename:  "test_task_compare_script_filename",
				CreatedAt: hashStringToTime("test_task_compare_script_created_at"),
				UpdatedAt: hashStringToTime("test_task_compare_script_updated_at"),
			},
			JudgeDir:           "",
			RunFilePath:        "",
			TimeUsed:           0,
			MemoryUsed:         0,
			OutputStrippedHash: "",
		}, task)
	})

	t.Run("NotFound", func(t *testing.T) {
		// Not Parallel
		taskStatus = "NotFound"
		_, err := GetTask()
		assert.Equal(t, ErrNotAvailable, err)
	})
}
