package web

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/suntt2019/EduOJJudger/base"
	"net/http"
)

func GetScript(name string) (err error) {
	httpResp, err := base.HC.R().Get(fmt.Sprintf("/script/%s", name))
	if err != nil {
		return errors.Wrap(err, "could not sent request")
	}
	if httpResp == nil {
		return errors.New("could not get response")
	}
	if httpResp.StatusCode() == http.StatusOK {
		return DownloadFile(httpResp, func(filename string) string {
			return viper.GetString("path.scripts") + "/" + name + "/" + filename
		})
	}
	err = HandleBackendErrorResponse(httpResp)
	if err != nil {
		return
	}
	err = HandleStorageErrorResponse(httpResp)
	if err != nil {
		return
	}
	log.Error(httpResp)
	return base.ErrUnknownTypeResponse
}
