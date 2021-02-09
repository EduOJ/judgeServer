package web

import (
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/suntt2019/EduOJJudger/base"
	"os"
	"path/filepath"
)

// GetFile download file from presigned url
func GetFile(location string, f func(filename string) string) error {
	resp, err := base.HC.R().Get(location)
	if err != nil {
		return errors.Wrap(err, "could not get presigned url response")
	}
	return DownloadFile(resp, f)
}

func DownloadFile(resp *resty.Response, f func(filename string) string) error {
	disposition := resp.Header().Get("Content-Disposition")
	// 18 is the length of `inline; filename="`
	if len(disposition) < 18 || disposition[:18] != `inline; filename="` || disposition[len(disposition)-1] != '"' {
		return errors.New("invalid content disposition")
	}
	fileName := disposition[18 : len(disposition)-1]
	filePath := f(fileName)
	if err := os.MkdirAll(filepath.Dir(filePath), 0777); err != nil { // TODO: use proper perm
		return errors.Wrap(err, "could not create folder")
	}
	file, err := os.Create(filePath)
	if err != nil {
		return errors.Wrap(err, "could not create file")
	}
	defer file.Close()
	content := resp.Body()
	if _, err := file.Write(content); err != nil {
		return errors.Wrap(err, "could not write file")
	}
	return nil
}
