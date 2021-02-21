package base

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestRemoveCache(t *testing.T) {
	// Not parallel
	assert.NoError(t, os.RemoveAll("../test_file/test_remove_buffer"))
	t.Cleanup(func() {
		_ = os.RemoveAll("../test_file/test_remove_buffer")
	})
	viper.Set("path.scripts", "../test_file/test_remove_buffer/scripts")
	viper.Set("path.test_cases", "../test_file/test_remove_buffer/test_cases")
	assert.NoError(t, os.MkdirAll(viper.GetString("path.scripts"), 0777))
	assert.NoError(t, os.MkdirAll(viper.GetString("path.test_cases"), 0777))
	assert.NoError(t, RemoveCache())
	_, err := os.Stat(viper.GetString("path.scripts"))
	assert.ErrorIs(t, err, os.ErrNotExist)
	_, err = os.Stat(viper.GetString("path.test_cases"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestIsFileLatest(t *testing.T) {
	t.Parallel()

	f, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	stat, err := os.Stat(f.Name())
	assert.NoError(t, err)

	t.Run("UpToDate", func(t *testing.T) {
		ok, err := IsFileLatest(f.Name(), stat.ModTime().Add(-1*time.Second))
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("Expired", func(t *testing.T) {
		ok, err := IsFileLatest(f.Name(), stat.ModTime().Add(time.Second))
		assert.NoError(t, err)
		assert.False(t, ok)
	})
}
