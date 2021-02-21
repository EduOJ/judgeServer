package api

import (
	"encoding/json"
	"fmt"
	"github.com/leoleoasd/EduOJBackend/app/request"
	"github.com/leoleoasd/EduOJBackend/app/response"
	"github.com/pkg/errors"
	"github.com/suntt2019/EduOJJudger/base"
	"io"
	"net/http"
	"strconv"
)

func UpdateRun(id uint, request *request.UpdateRunRequest, runFile, buildOutputFile, compareOutputFile io.Reader) error {
	req := base.HttpClient.R().SetMultipartFormData(map[string]string{
		"status":               request.Status,
		"memory_used":          strconv.Itoa(int(request.MemoryUsed)),
		"time_used":            strconv.Itoa(int(request.TimeUsed)),
		"output_stripped_hash": request.OutputStrippedHash,
		"message":              request.Message,
	}).
		SetMultipartField("OutputFile", "OutputFile", "application/octet-stream", runFile).
		SetMultipartField("CompilerFile", "CompilerFile", "application/octet-stream", buildOutputFile).
		SetMultipartField("ComparerFile", "ComparerFile", "application/octet-stream", compareOutputFile)

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
