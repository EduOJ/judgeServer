package api

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"path"
	"testing"
)

func script(wr http.ResponseWriter, r *http.Request, uri string) {
	http.Redirect(wr, r, path.Join("/fileURI", uri+".zip", uri+"_content"), http.StatusFound)
}

func TestGetScript(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		f, err := GetScript("test_get_script_success")
		assert.NoError(t, err)
		body, err := ioutil.ReadAll(f)
		assert.NoError(t, err)
		assert.Equal(t, "test_get_script_success_content", string(body))
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		_, err := GetScript("test_get_script_NON_EXISTING")
		assert.NotNil(t, err)
		assert.Equal(t, `unexpected response: <?xml version="1.0" encoding="UTF-8"?>`+
			"<Error><Code>NoSuchKey</Code>"+
			"<Message>The specified key does not exist.</Message>"+
			"</Error>", err.Error())
	})
}
