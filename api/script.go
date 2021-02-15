package api

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/suntt2019/EduOJJudger/base"
	"path"
)

func GetScript(name string) error {
	_, err := base.HttpClient.R().SetOutput(path.Join(viper.GetString("path.scripts"), "downloads", name+".zip")).
		Get(fmt.Sprintf("/script/%s", name))
	if err != nil {
		return errors.Wrap(err, "could not sent request")
	}
	return nil
}
