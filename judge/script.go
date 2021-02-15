package judge

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/suntt2019/EduOJJudger/api"
	"github.com/suntt2019/EduOJJudger/base"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"
)

// check if the script is expired or non-existed then update it
func EnsureLatestScript(name string, updatedAt time.Time) error {
	ok, err := CheckScript(name, updatedAt)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	dir, err := api.GetScript(name)
	if err != nil {
		return err
	}
	return InstallScript(name, dir)
}

// CheckScript check update time of script folder, if the script need to update, it returns false
func CheckScript(name string, updatedAt time.Time) (ok bool, err error) {
	scriptsPath := viper.GetString("path.scripts")
	info, err := os.Stat(path.Join(scriptsPath, name))
	if err != nil && !os.IsNotExist(err) {
		return true, errors.Wrap(err, "unexpected stat of script directory")
	}
	// check if expired or non-exist
	if os.IsNotExist(err) || !info.IsDir() || info.ModTime().Before(updatedAt) {
		return false, nil
	}
	return true, nil
}

func InstallScript(name string, dir string) error {
	scriptsPath := viper.GetString("path.scripts")

	// remove expired version
	if err := os.RemoveAll(path.Join(scriptsPath, name)); err != nil {
		return errors.Wrap(err, "could not remove directory")
	}

	// create directory
	if err := os.MkdirAll(path.Join(scriptsPath, name), 0777); err != nil {
		return errors.Wrap(err, "could not remove directory")
	}
	if err := base.ScriptUser.OwnRWX(path.Join(scriptsPath, name)); err != nil {
		return errors.Wrap(err, "could not set mod for directory")
	}

	// unzip
	err := base.ScriptUser.RunWithTimeout(exec.Command(
		"unzip", path.Join(dir, name+".zip"), "-d", path.Join(scriptsPath, name)), viper.GetDuration("timeout.script.unzip"))
	if err != nil {
		return errors.Wrap(err, "could not unzip script zip file")
	}

	// set perm for compile file
	if err := base.ScriptUser.OwnRWX(path.Join(scriptsPath, name, "compile")); err != nil {
		return errors.Wrap(err, "could not set permission for script directory")
	}

	// compile
	cmd := exec.Command("./compile")
	cmd.Dir = path.Join(scriptsPath, name)
	err = base.ScriptUser.RunWithTimeout(cmd, viper.GetDuration("timeout.script.compile"))
	if err != nil {
		return errors.Wrap(err, "could not compile script")
	}

	// set permission for run
	if err := base.ScriptUser.OwnRWX(path.Join(scriptsPath, name)); err != nil {
		return errors.Wrap(err, "could not set permission for script run file")
	}

	return nil
}

func RunScript(name string, dir string, args ...string) error {
	absPath, err := filepath.Abs(path.Join(viper.GetString("path.scripts"), name, "run"))
	if err != nil {
		return err
	}
	cmd := exec.Command(absPath, args...)
	cmd.Dir = dir
	return base.ScriptUser.Run(cmd)
}

func RunScriptWithTimeout(timeout time.Duration, name string, dir string, args ...string) error {
	return base.WithTimeout(timeout, func() error {
		return RunScript(name, dir, args...)
	})
}
