package judge

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/suntt2019/EduOJJudger/api"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"
)

var ScriptLock sync.Mutex

func RunScript(name string, latestUpdateTime time.Time) (err error) {
	err = ensureLatestScript(name, latestUpdateTime)
	if err != nil {
		return
	}
	return runScript(name)
}

func ensureLatestScript(name string, latestUpdateTime time.Time) error {
	ScriptLock.Lock()
	defer ScriptLock.Unlock()
	ok, err := checkScript(name, latestUpdateTime)
	if err != nil {
		return err
	}
	if !ok {
		f, err := api.GetScript(name)
		if err != nil {
			return err
		}
		if err = installScript(name, f); err != nil {
			return err
		}
	}
	return nil
}

// checkScript checks if the script is the latest version, it returns true if the script is the latest version.
func checkScript(name string, latestUpdateTime time.Time) (ok bool, err error) {
	stat, err := os.Stat(path.Join(viper.GetString("path.scripts"), name))
	if os.IsNotExist(err) || !stat.IsDir() || stat.ModTime().Before(latestUpdateTime) {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "could not check stat for script")
	}
	return true, nil
}

// installScript unzips temped zip file and compiles the script.
func installScript(name string, file *os.File) error {

	err := os.RemoveAll(path.Join(viper.GetString("path.scripts"), name))
	if err != nil {
		return errors.Wrap(err, "could not remove old version script")
	}

	// TODO: Fix issue
	err = exec.Command("unzip", file.Name(), "-d", path.Join(viper.GetString("path.scripts"), name)).Run()
	if err != nil {
		return errors.Wrap(err, "could not unzip script zip file")
	}

	cmd := exec.Command("./compile")
	cmd.Dir = path.Join(viper.GetString("path.scripts"), name)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "could not compile script")
	}

	return nil
}

func runScript(name string) error {
	cmd := exec.Command("./run")
	cmd.Dir = path.Join(viper.GetString("path.scripts"), name)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "could not run script")
	}

	return nil
}
