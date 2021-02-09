package web_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/suntt2019/EduOJJudger/base"
	"github.com/suntt2019/EduOJJudger/web"
	"net/http"
	"testing"
)

func script(wr http.ResponseWriter, r *http.Request) {
	// r.RequestURI[8:] remove "/script/"
	name := r.RequestURI[8:]
	switch name {
	case "non_existing_script":
		http.Redirect(wr, r, "/backendErr/404/NOT_FOUND", http.StatusFound)
	default:
		http.Redirect(wr, r, "/fileURI/script_"+name+".zip/script_"+name+"_content", http.StatusFound)
	}
}

func TestGetScript(t *testing.T) {
	t.Parallel()

	t.Run("Success", func(t *testing.T) {
		assert.Nil(t, web.GetScript("test_get_script_success"))
		checkFile(t, "../test_file/scripts/test_get_script_success/script_test_get_script_success.zip", "script_test_get_script_success_content")
	})

	t.Run("NotFound", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, base.ErrNotFoundResponse, web.GetScript("non_existing_script"))
	})
}
