package base

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestRemoveBuffer(t *testing.T) {
	// Not parallel
	assert.Nil(t, os.RemoveAll("../test_file/test_remove_buffer"))
	t.Cleanup(func() {
		_ = os.RemoveAll("../test_file/test_remove_buffer")
	})
	ScriptPath = "../test_file/test_remove_buffer/scripts"
	RunPath = "../test_file/test_remove_buffer/runs"
	assert.Nil(t, os.MkdirAll(ScriptPath, 0777))
	assert.Nil(t, os.MkdirAll(RunPath, 0777))
	assert.Nil(t, RemoveBuffer())
	_, err := os.Stat(ScriptPath)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(RunPath)
	assert.True(t, os.IsNotExist(err))
}
