package api

import (
	"github.com/EduOJ/judgeServer/base"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
)

func GetFile(url string, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "could not create file")
	}
	resp, err := base.HttpClient.R().SetOutput(path).Get(url)
	if err != nil {
		return errors.Wrap(err, "could not send request")
	}
	if resp.StatusCode() == http.StatusOK {
		return nil
	}
	body, err := ioutil.ReadAll(f)
	if err != nil {
		return errors.Wrap(err, "could not read response body")
	}
	return errors.New("unexpected response: " + string(body))
}
