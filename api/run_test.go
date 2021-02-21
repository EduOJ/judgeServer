package api

import (
	"github.com/leoleoasd/EduOJBackend/app/request"
	"github.com/leoleoasd/EduOJBackend/app/response"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

type fakeRun struct {
	ID     uint
	Status string // READY / ANOTHER / JUDGED
	Form   *multipart.Form
}

// 1: Success
// 2: WrongRunID
var runs = make(map[uint64]*fakeRun, 2)

func run(wr http.ResponseWriter, r *http.Request, uri string) {
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

	rr := multipart.NewReader(r.Body, r.Header.Get("Content-Type")[30:])
	form, err := rr.ReadForm(1024 * 1024 * 10)
	if err != nil {
		panic(err)
	}
	runs[runID].Form = form
	if err = marshalAndWrite(wr, resp); err != nil {
		panic(err)
	}
}

func checkMultipartFile(t *testing.T, fileHeader *multipart.FileHeader, fileName, content string) {
	assert.Equal(t, fileName, fileHeader.Filename)
	assert.Equal(t, `form-data; name="`+fileName+`"; filename="`+fileName+`"`, fileHeader.Header.Get("Content-Disposition"))
	assert.Equal(t, "application/octet-stream", fileHeader.Header.Get("Content-Type"))
	f, err := fileHeader.Open()
	assert.NoError(t, err)
	b, err := ioutil.ReadAll(f)
	assert.NoError(t, err)
	assert.Equal(t, content, string(b))
}

func TestUpdateRun(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		// Not Parallel
		runs[1] = &fakeRun{
			ID:     1,
			Status: "READY",
		}
		req := request.UpdateRunRequest{
			Status:             "ACCEPTED",
			MemoryUsed:         1024,
			TimeUsed:           1000,
			OutputStrippedHash: "test_update_run_success_output_hash",
			Message:            "test_update_run_success_output_message",
		}
		err := UpdateRun(1, &req,
			strings.NewReader("test_update_run_success_run_file"),
			strings.NewReader("test_update_run_success_build_output"),
			strings.NewReader("test_update_run_success_compare_output"))
		assert.NoError(t, err)
		expectedFormValue := map[string][]string{
			"memory_used":          {"1024"},
			"message":              {"test_update_run_success_output_message"},
			"output_stripped_hash": {"test_update_run_success_output_hash"},
			"status":               {"ACCEPTED"},
			"time_used":            {"1000"},
		}
		checkMultipartFile(t, runs[1].Form.File["OutputFile"][0], "OutputFile", "test_update_run_success_run_file")
		checkMultipartFile(t, runs[1].Form.File["CompilerFile"][0], "CompilerFile", "test_update_run_success_build_output")
		checkMultipartFile(t, runs[1].Form.File["ComparerFile"][0], "ComparerFile", "test_update_run_success_compare_output")
		assert.Equal(t, expectedFormValue, runs[1].Form.Value)
	})

	t.Run("WrongRunID", func(t *testing.T) {
		// Not Parallel
		runs[2] = &fakeRun{
			ID:     1,
			Status: "ANOTHER",
		}
		req := request.UpdateRunRequest{
			Status:             "WRONG_ANSWER",
			MemoryUsed:         2048,
			TimeUsed:           2000,
			OutputStrippedHash: "test_update_run_another_run_id_output_hash",
			Message:            "test_update_run_another_run_id_output_message",
		}
		err := UpdateRun(2, &req,
			strings.NewReader("test_update_run_another_run_id_run_file"),
			strings.NewReader("test_update_run_another_run_id_build_output"),
			strings.NewReader("test_update_run_another_run_id_compare_output"))
		assert.NotNil(t, err)
		assert.Equal(t, `unexpected response: {"message":"WRONG_RUN_ID","error":null,"data":null}`, err.Error())
		expectedFormValue := map[string][]string{
			"memory_used":          {"2048"},
			"message":              {"test_update_run_another_run_id_output_message"},
			"output_stripped_hash": {"test_update_run_another_run_id_output_hash"},
			"status":               {"WRONG_ANSWER"},
			"time_used":            {"2000"},
		}
		checkMultipartFile(t, runs[2].Form.File["OutputFile"][0], "OutputFile", "test_update_run_another_run_id_run_file")
		checkMultipartFile(t, runs[2].Form.File["CompilerFile"][0], "CompilerFile", "test_update_run_another_run_id_build_output")
		checkMultipartFile(t, runs[2].Form.File["ComparerFile"][0], "ComparerFile", "test_update_run_another_run_id_compare_output")
		assert.Equal(t, expectedFormValue, runs[2].Form.Value)
	})
}
