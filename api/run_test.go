package api

import (
	"github.com/leoleoasd/EduOJBackend/app/request"
	"github.com/leoleoasd/EduOJBackend/app/response"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strconv"
	"testing"
)

type fakeRun struct {
	ID            uint
	Status        string // READY / ANOTHER / JUDGED
	UpdateRequest request.UpdateRunRequest
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
	req := request.UpdateRunRequest{}
	if err = readAndUnmarshal(r.Body, &req); err != nil {
		panic(err)
	}
	runs[runID].UpdateRequest = req
	if err = marshalAndWrite(wr, resp); err != nil {
		panic(err)
	}
}

func TestUpdateRun(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		// Not Parallel
		runs[1] = &fakeRun{
			ID:            1,
			Status:        "READY",
			UpdateRequest: request.UpdateRunRequest{},
		}
		req := request.UpdateRunRequest{
			Status:             "ACCEPTED",
			MemoryUsed:         1024,
			TimeUsed:           1000,
			OutputStrippedHash: "test_update_run_success_output_hash",
			Message:            "test_update_run_success_output_message",
		}
		err := UpdateRun(1, &req)
		assert.NoError(t, err)
		assert.Equal(t, req, runs[1].UpdateRequest)
	})

	t.Run("WrongRunID", func(t *testing.T) {
		// Not Parallel
		runs[2] = &fakeRun{
			ID:            1,
			Status:        "ANOTHER",
			UpdateRequest: request.UpdateRunRequest{},
		}
		req := request.UpdateRunRequest{
			Status:             "WRONG_ANSWER",
			MemoryUsed:         2048,
			TimeUsed:           2000,
			OutputStrippedHash: "test_update_run_another_run_id_output_hash",
			Message:            "test_update_run_another_run_id_output_message",
		}
		err := UpdateRun(2, &req)
		assert.NotNil(t, err)
		assert.Equal(t, `unexpected response: {"message":"WRONG_RUN_ID","error":null,"data":null}`, err.Error())
		assert.Equal(t, req, runs[2].UpdateRequest)
	})
}
