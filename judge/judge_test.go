package judge

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/leoleoasd/EduOJBackend/app/request"
	"github.com/leoleoasd/EduOJBackend/app/response"
	"github.com/leoleoasd/EduOJBackend/database/models"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/suntt2019/EduOJJudger/base"
	"hash/fnv"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func createWithDir(name string) (*os.File, error) {
	if err := os.MkdirAll(path.Dir(name), 0777); err != nil { // TODO: perm
		return nil, err
	}
	return os.Create(name)
}

func createAndWrite(name, content string) error {
	f, err := createWithDir(name)
	if err != nil {
		return err
	}
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	return f.Close()
}

const maxIdCount = 200

var testIdLock sync.Mutex
var testIdMap = make(map[uint]string, maxIdCount)

func hashStringToId(s string) uint {
	testIdLock.Lock()
	defer testIdLock.Unlock()
	h := fnv.New32()
	if _, err := h.Write([]byte(s)); err != nil {
		panic(err)
	}
	id := uint(h.Sum32()) % maxIdCount
	crashTimes := 0
	for testIdMap[id] != "" {
		id++
		crashTimes++
		if crashTimes == maxIdCount {
			panic("test id map is full")
		}
		if id == maxIdCount {
			id = 0
		}
	}
	testIdMap[id] = s
	return id
}

func hashStringToTime(s string) time.Time {
	h := fnv.New32()
	if _, err := h.Write([]byte(s)); err != nil {
		panic(err)
	}
	return time.Unix(int64(h.Sum32()), 0).UTC()
}

func readAndUnmarshal(reader io.Reader, out interface{}) error {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func marshalAndWrite(writer io.Writer, in interface{}) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	_, err = writer.Write(b)
	return err
}

func testServerRoute(wr http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.RequestURI)
	if err != nil {
		panic(err)
	}
	index := strings.Index(u.Path[1:], "/")
	var service, uri string
	if index == -1 {
		service = u.Path[1:]
		uri = ""
	} else {
		service = u.Path[1 : index+1]
		uri = u.Path[index+2:]
	}
	switch service {
	case "echoURI":
		echoURI(wr, uri)
	case "fileURI":
		fileURI(wr, uri)
	case "script":
		script(wr, r, uri)
	case "task":
		task(wr, u.Query())
	case "run":
		runURI(wr, r, uri)
	default:
		panic(`invalid service for test server: "` + service + `"`)
	}
}

func echoURI(wr http.ResponseWriter, uri string) {
	if _, err := wr.Write([]byte(uri)); err != nil {
		panic(err)
	}
}

func fileURI(wr http.ResponseWriter, uri string) {
	content := strings.Split(uri, "/")
	if len(content) != 2 {
		panic("unexpected content count")
	}
	if strings.Contains(content[0], "NON_EXISTING") {
		wr.WriteHeader(http.StatusNotFound)
		_, err := wr.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>` +
			"<Error><Code>NoSuchKey</Code>" +
			"<Message>The specified key does not exist.</Message>" +
			"</Error>"))
		if err != nil {
			panic(errors.Wrap(err, "could not write response"))
		}
		return
	}
	wr.Header().Set("Content-Disposition", `inline; filename="`+content[0]+`"`) // filename
	if _, err := wr.Write([]byte(content[1])); err != nil {
		panic(err)
	}
}

func script(wr http.ResponseWriter, r *http.Request, uri string) {
	http.Redirect(wr, r, path.Join("/fileURI", uri+".zip", uri+"_content"), http.StatusFound)
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
				RunID              uint            `json:"run_id"`
				Language           models.Language `json:"language"`
				TestCaseID         uint            `json:"test_case_id"`
				InputFile          string          `json:"input_file"`
				OutputFile         string          `json:"output_file"`
				CodeFile           string          `json:"code_file"`
				TestCaseUpdatedAt  time.Time       `json:"test_case_updated_at"`
				MemoryLimit        uint64          `json:"memory_limit"`
				TimeLimit          uint            `json:"time_limit"`
				CompileEnvironment string          `json:"compile_environment"`
				CompareScript      models.Script   `json:"compare_script"`
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
				TestCaseID:         1,
				InputFile:          "test_task_input_file_presigned_url",
				OutputFile:         "test_task_output_file_presigned_url",
				CodeFile:           "test_task_code_file_presigned_url",
				TestCaseUpdatedAt:  hashStringToTime("test_task_test_case_updated_at"),
				MemoryLimit:        1024,
				TimeLimit:          1000,
				CompileEnvironment: "test,task,build,args",
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

type fakeRun struct {
	ID            uint
	Status        string // READY / ANOTHER / JUDGED
	UpdateRequest request.UpdateRunRequest
}

var runs = make(map[uint64]*fakeRun, 2)

func runURI(wr http.ResponseWriter, r *http.Request, uri string) {
	runID, err := strconv.ParseUint(uri, 10, 64)
	if err != nil {
		panic(err)
	}
	resp := response.Response{
		Message: "",
		Error:   nil,
		Data:    nil,
	}
	switch runs[runID].Status {
	case "READY":
		resp.Message = "SUCCESS"
	case "ANOTHER":
		wr.WriteHeader(http.StatusForbidden)
		resp.Message = "WRONG_RUN_ID"
	case "JUDGED":
		wr.WriteHeader(http.StatusBadRequest)
		resp.Message = "ALREADY_SUBMITTED"
	default:
		panic("Unexpected run status")
	}
	req := request.UpdateRunRequest{}
	if err = readAndUnmarshal(r.Body, &req); err != nil {
		panic(err)
	}
	runs[runID].UpdateRequest = req
	if err = marshalAndWrite(wr, resp); err != nil {
		panic(err)
	}
}

func checkFile(t *testing.T, path, content string) {
	file, err := os.Open(path)
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(file)
	assert.NoError(t, err)
	assert.Equal(t, content, string(b))
	err = file.Close()
	assert.NoError(t, err)
}

func checkFileNonExist(t *testing.T, path string) {
	_, err := os.OpenFile(path, os.O_RDONLY, 0666)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestMain(m *testing.M) {
	config := `user:
  compile: build_user
  run: run_user
path:
  scripts: ../test_file/scripts
  test_cases: ../test_file/test_cases
  temp: ../test_file/temp
judge:
  build:
    max_time: 1s
    max_memory: 104857600
    max_stack: 104857600
    max_output_size: 104857600
  run:
    max_output_size: 104857600
`
	viper.SetConfigType("yml")
	if err := viper.ReadConfig(strings.NewReader(config)); err != nil {
		panic(err)
	}
	dir, err := ioutil.TempDir("", "eduoj_judger_test_scripts_*")
	if err != nil {
		panic(errors.Wrap(err, "could not create temp scripts dir"))
	}
	viper.Set("path.scripts", dir)
	dir, err = ioutil.TempDir("", "eduoj_judger_test_test_cases_*")
	if err != nil {
		panic(errors.Wrap(err, "could not create temp test cases dir"))
	}
	viper.Set("path.test_cases", dir)
	l, err := ioutil.TempFile("", "eduoj_judger_test_log_*")
	if err != nil {
		panic(errors.Wrap(err, "could not create temp log file"))
	}
	viper.Set("log.sandbox_log_path", l.Name())
	ts := httptest.NewServer(http.HandlerFunc(testServerRoute))
	base.HttpClient = resty.New().SetHostURL(ts.URL)
	ret := m.Run()
	fmt.Println("test id map:")
	for id, name := range testIdMap {
		if name != "" {
			fmt.Printf("#%3d -> %s\n", id, name)
		}
	}
	os.Exit(ret)
}
