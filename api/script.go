package api

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
)

// To ensure the atomicity of updating script, this function SHOULD NOT
// be used in functions other than judge.EnsureLatestScript.
func GetScript(name string) (*os.File, error) {
	f, err := ioutil.TempFile("", fmt.Sprintf("eduoj_judger_script_%s_*", name))
	if err != nil {
		return f, errors.Wrap(err, "could not create temp file")
	}
	err = GetFile(path.Join("script", name), f.Name())
	if err != nil {
		return nil, errors.Wrap(err, "could get file")
	}
	return f, nil
}
