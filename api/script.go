package api

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/suntt2019/EduOJJudger/base"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

// To ensure the atomicity of updating script, this function SHOULD NOT
// be used in functions other than judge.EnsureLatestScript.
func GetScript(name string) (*os.File, error) {
	f, err := ioutil.TempFile("", fmt.Sprintf("eduoj_judger_script_%s_*", name))
	if err != nil {
		return f, errors.Wrap(err, "could not create temp file")
	}
	resp, err := base.HttpClient.R().SetOutput(f.Name()).
		Get(path.Join("script", name))
	if err != nil {
		return nil, errors.Wrap(err, "could not send request")
	}
	if resp.StatusCode() == http.StatusOK {
		return f, nil
	}
	body, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "could not read response body")
	}
	return nil, errors.New("unexpected response: " + string(body))
}
