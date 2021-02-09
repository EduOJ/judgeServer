package web_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/suntt2019/EduOJJudger/web"
	"net/http"
	"os"
	"strings"
	"testing"
)

func fileURI(wr http.ResponseWriter, r *http.Request) {
	// r.RequestURI[9:] remove "/fileURI/"
	content := strings.Split(r.RequestURI[9:], "/")
	if len(content) != 2 {
		panic("unexpected content count")
	}
	wr.Header().Set("Content-Disposition", `inline; filename="`+content[0]+`"`) // filename
	if _, err := wr.Write([]byte(content[1])); err != nil {
		panic(err)
	}
}

func TestGetFile(t *testing.T) {
	t.Parallel()
	assert.Nil(t, os.RemoveAll("../test_file/test_get_file"))
	assert.Nil(t, os.MkdirAll("../test_file/test_get_file", 0777))
	t.Cleanup(func() {
		_ = os.RemoveAll("../test_file/test_get_file")
	})
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, web.GetFile("/fileURI/success_file_name.txt/success_file_content", func(filename string) string {
			return "../test_file/test_get_file/" + filename
		}))
		checkFile(t, "../test_file/test_get_file/success_file_name.txt", "success_file_content")
	})
	t.Run("WrongContentDisposition", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "invalid content disposition",
			web.GetFile("/echoURI/test_get_file", func(filename string) string {
				return "../test_file/test_get_file/" + filename
			}).Error())
	})
}
