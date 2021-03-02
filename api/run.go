package api

import (
	"encoding/json"
	"fmt"
	"github.com/EduOJ/backend/app/request"
	"github.com/EduOJ/backend/app/response"
	"github.com/EduOJ/judgeServer/base"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"strconv"
)

func UpdateRun(id uint, request *request.UpdateRunRequest, runFile, buildOutputFile, compareOutputFile io.Reader) error {
	req := base.HttpClient.R().SetMultipartFormData(map[string]string{
		"status":               request.Status,
		"memory_used":          strconv.Itoa(int(*request.MemoryUsed)),
		"time_used":            strconv.Itoa(int(*request.TimeUsed)),
		"output_stripped_hash": *request.OutputStrippedHash,
		"message":              request.Message,
	}).
		SetMultipartField("output_file", "output_file", "application/octet-stream", runFile).
		SetMultipartField("compiler_output_file", "compiler_output_file", "application/octet-stream", buildOutputFile).
		SetMultipartField("comparer_output_file", "comparer_output_file", "application/octet-stream", compareOutputFile)

	httpResp, err := req.Put(fmt.Sprintf("run/%d", id))
	if err != nil {
		return errors.Wrap(err, "could not send put request")
	}
	resp := response.Response{}
	if err = json.Unmarshal(httpResp.Body(), &resp); err != nil {
		return errors.Wrap(err, "could not unmarshal response body")
	}
	if httpResp.StatusCode() == http.StatusOK && resp.Message == "SUCCESS" {
		return nil
	}
	return errors.New("unexpected response: " + string(httpResp.Body()))
}
