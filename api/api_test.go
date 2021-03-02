package api

import (
	"encoding/json"
	"github.com/EduOJ/judgeServer/base"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"hash/fnv"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func hashStringToTime(s string) time.Time {
	h := fnv.New32()
	if _, err := h.Write([]byte(s)); err != nil {
		panic(err)
	}
	return time.Unix(int64(h.Sum32()), 0).UTC()
}

func readAndUnmarshal(reader io.Reader, out interface{}) error {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func marshalAndWrite(writer io.Writer, in interface{}) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}
	_, err = writer.Write(b)
	return err
}

func testServerRoute(wr http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.RequestURI)
	if err != nil {
		panic(err)
	}
	index := strings.Index(u.Path[1:], "/")
	var service, uri string
	if index == -1 {
		service = u.Path[1:]
		uri = ""
	} else {
		service = u.Path[1 : index+1]
		uri = u.Path[index+2:]
	}
	switch service {
	case "echoURI":
		echoURI(wr, uri)
	case "fileURI":
		fileURI(wr, uri)
	case "script":
		script(wr, r, uri)
	case "task":
		task(wr, u.Query())
	case "run":
		run(wr, r, uri)
	default:
		panic(`invalid service for test server: "` + service + `"`)
	}
}

func echoURI(wr http.ResponseWriter, uri string) {
	if _, err := wr.Write([]byte(uri)); err != nil {
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
	viper.Set("path.test_cases", "../test_file/test_cases")
	os.Exit(m.Run())
}
