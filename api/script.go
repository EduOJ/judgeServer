package api

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/suntt2019/EduOJJudger/base"
	"io/ioutil"
	"path"
)

func GetScript(name string) (string, error) {
	dir, err := ioutil.TempDir(viper.GetString("path.temp"), "")
	if err != nil {
		return "", errors.Wrap(err, "could not get temp dir")
	}
	if err := base.ScriptUser.OwnRWX(dir); err != nil {
		return "", errors.Wrap(err, "could not set permission for temp dir")
	}
	_, err = base.HttpClient.R().SetOutput(path.Join(dir, name+".zip")).
		Get(fmt.Sprintf("script/%s", name))
	if err != nil {
		return "", errors.Wrap(err, "could not sent request")
	}
	return dir, nil
}
