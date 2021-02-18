package api

import (
	"encoding/json"
	"fmt"
	"github.com/leoleoasd/EduOJBackend/database/models"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/suntt2019/EduOJJudger/base"
	judger "github.com/suntt2019/Judger"
	"net/http"
	"os"
	"path"
	"time"
)

type Task struct {
	RunID              uint            `json:"run_id"`
	Language           models.Language `json:"language"`
	TestCaseID         uint            `json:"test_case_id"`
	InputFile          string          `json:"input_file"`  // pre-signed url
	OutputFile         string          `json:"output_file"` // same as above
	InputFilePath      string
	OutputFilePath     string
	CodeFile           string        `json:"code_file"`
	TestCaseUpdatedAt  time.Time     `json:"test_case_updated_at"`
	MemoryLimit        uint64        `json:"memory_limit"`        // Byte
	TimeLimit          uint          `json:"time_limit"`          // ms
	BuildArg           string        `json:"compile_environment"` // E.g.  O2=false
	CompareScript      models.Script `json:"compare_script"`
	JudgeDir           string
	RunFile            *os.File
	TimeUsed           uint
	MemoryUsed         uint
	OutputStrippedHash string
	CompareResult      bool
	JudgeResult        judger.JudgeResult
}

func (t *Task) GenerateFilePath() {
	t.InputFilePath = path.Join(viper.GetString("path.test_cases"), fmt.Sprintf("%d", t.TestCaseID), "in")
	t.OutputFilePath = path.Join(viper.GetString("path.test_cases"), fmt.Sprintf("%d", t.TestCaseID), "out")
}

type getTaskResponse struct {
	Message string      `json:"message"`
	Error   interface{} `json:"error"`
	Data    Task        `json:"data"`
}

func GetTask() (*Task, error) {
	httpResp, err := base.HttpClient.R().SetQueryParam("poll", "1").Get("task")
	if err != nil {
		return nil, errors.Wrap(err, "could not send get request")
	}
	resp := getTaskResponse{}
	if err = json.Unmarshal(httpResp.Body(), &resp); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal response")
	}
	if httpResp.StatusCode() == http.StatusNotFound && resp.Message == "NOT_FOUND" {
		return nil, base.ErrNotAvailable
	}
	if httpResp.StatusCode() == http.StatusOK && resp.Message == "SUCCESS" {
		resp.Data.GenerateFilePath()
		return &resp.Data, nil
	}
	return nil, errors.New("unexpected response: " + string(httpResp.Body()))
}
