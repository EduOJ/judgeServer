package base

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestRemoveCache(t *testing.T) {
	// Not parallel
	assert.NoError(t, os.RemoveAll("../test_file/test_remove_buffer"))
	t.Cleanup(func() {
		_ = os.RemoveAll("../test_file/test_remove_buffer")
	})
	viper.Set("path.scripts", "../test_file/test_remove_buffer/scripts")
	viper.Set("path.runs", "../test_file/test_remove_buffer/runs")
	assert.NoError(t, os.MkdirAll(viper.GetString("path.scripts"), 0777))
	assert.NoError(t, os.MkdirAll(viper.GetString("path.runs"), 0777))
	assert.NoError(t, RemoveCache())
	_, err := os.Stat(viper.GetString("path.scripts"))
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(viper.GetString("path.runs"))
	assert.True(t, os.IsNotExist(err))
}
