package base

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os"
	"time"
)

func RemoveCache() error {
	err := os.RemoveAll(viper.GetString("path.scripts"))
	if err != nil {
		return err
	}
	err = os.RemoveAll(viper.GetString("path.test_cases"))
	if err != nil {
		return err
	}
	return nil
}

func IsFileLatest(path string, latestUpdateTime time.Time) (ok bool, err error) {
	stat, err := os.Stat(path)
	if os.IsNotExist(err) || stat.IsDir() || stat.ModTime().Before(latestUpdateTime) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "could not check stat for script")
	}
	return true, nil
}
