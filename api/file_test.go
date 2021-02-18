package api

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"testing"
)

func fileURI(wr http.ResponseWriter, uri string) {
	content := strings.Split(uri, "/")
	if len(content) != 2 {
		panic("unexpected content count")
	}
	if strings.Contains(content[0], "NON_EXISTING") {
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

func TestGetFile(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		file, err := ioutil.TempFile("", "eduoj_judger_test_get_file_*")
		assert.NoError(t, err)
		err = GetFile(path.Join("/fileURI", "test_get_file_success", "test_get_file_success_content"), file.Name())
		assert.NoError(t, err)
		b, err := ioutil.ReadAll(file)
		assert.NoError(t, err)
		assert.Equal(t, "test_get_file_success_content", string(b))
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		file, err := ioutil.TempFile("", "eduoj_judger_test_get_file_*")
		assert.NoError(t, err)
		err = GetFile(path.Join("/fileURI", "test_get_file_NON_EXISTING", "non_existing_content"), file.Name())
		assert.NotNil(t, err)
		assert.Equal(t, `unexpected response: <?xml version="1.0" encoding="UTF-8"?>`+
			"<Error><Code>NoSuchKey</Code>"+
			"<Message>The specified key does not exist.</Message>"+
			"</Error>", err.Error())
	})
}
