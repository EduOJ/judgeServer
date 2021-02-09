package base

import (
	"github.com/spf13/viper"
	"os"
)

func RemoveCache() error {
	err := os.RemoveAll(viper.GetString("path.scripts"))
	if err != nil {
		return err
	}
	err = os.RemoveAll(viper.GetString("path.runs"))
	if err != nil {
		return err
	}
	return nil
}
