package api_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/suntt2019/EduOJJudger/api"
	"net/http"
	"os"
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
	assert.Nil(t, os.RemoveAll("../test_file/scripts/downloads/test_get_script_success.zip"))
	t.Cleanup(func() {
		assert.Nil(t, os.RemoveAll("../test_file/scripts/downloads/test_get_script_success.zip"))
	})
	assert.Nil(t, api.GetScript("test_get_script_success"))
	checkFile(t, "../test_file/scripts/downloads/test_get_script_success.zip", "script_test_get_script_success_content")
}
