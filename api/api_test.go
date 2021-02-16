package api

import (
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/suntt2019/EduOJJudger/base"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func testServerRoute(wr http.ResponseWriter, r *http.Request) {
	index := strings.Index(r.RequestURI[1:], "/")
	if index == -1 {
		panic("could not find the second '/' to find out service name")
	}
	uri := r.RequestURI[index+2:]
	switch r.RequestURI[1 : index+1] {
	case "echoURI":
		echoURI(wr, uri)
	case "fileURI":
		fileURI(wr, uri)
	case "script":
		script(wr, r, uri)
	default:
		panic(`invalid service for test server: "` + r.RequestURI[1:index+1] + `"`)
	}

}

func echoURI(wr http.ResponseWriter, uri string) {
	if _, err := wr.Write([]byte(uri)); err != nil {
		panic(err)
	}
}

func fileURI(wr http.ResponseWriter, uri string) {
	content := strings.Split(uri, "/")
	if len(content) != 2 {
		panic("unexpected content count")
	}
	if strings.Contains(content[1], "NON_EXISTING") {
		wr.WriteHeader(http.StatusNotFound)
		_, err := wr.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>` +
			"<Error><Code>NoSuchKey</Code>" +
			"<Message>The specified key does not exist.</Message>" +
			"</Error>"))
		if err != nil {
			panic(errors.Wrap(err, "could not write response"))
		}
		return
	}
	wr.Header().Set("Content-Disposition", `inline; filename="`+content[0]+`"`) // filename
	if _, err := wr.Write([]byte(content[1])); err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	config := ``
	if err := viper.ReadConfig(strings.NewReader(config)); err != nil {
		panic(err)
	}
	ts := httptest.NewServer(http.HandlerFunc(testServerRoute))
	base.HttpClient = resty.New().SetHostURL(ts.URL)
	viper.Set("path.scripts", "../test_file/scripts")
	viper.Set("path.runs", "../test_file/runs")
	ret := m.Run()
	os.Exit(ret)
}
