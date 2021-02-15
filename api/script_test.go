package api_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/suntt2019/EduOJJudger/api"
	"net/http"
	"path"
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
	dir, err := api.GetScript("test_get_script_success")
	assert.NoError(t, err)
	checkFile(t, path.Join(dir, "test_get_script_success.zip"), "script_test_get_script_success_content")
}
