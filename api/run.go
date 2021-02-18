package api

import (
	"encoding/json"
	"github.com/leoleoasd/EduOJBackend/app/request"
	"github.com/leoleoasd/EduOJBackend/app/response"
	"github.com/pkg/errors"
	"github.com/suntt2019/EduOJJudger/base"
	"net/http"
	"path"
	"strconv"
)

func UpdateRun(id uint, request *request.UpdateRunRequest) error {
	httpResp, err := base.HttpClient.R().SetBody(request).Put(path.Join("run", strconv.Itoa(int(id))))
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
