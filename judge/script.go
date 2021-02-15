package judge

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/suntt2019/EduOJJudger/base"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// check if the script is expired or uninstalled and update it
func InstallScript(name string, updatedAt time.Time) error {
	scriptsPath := viper.GetString("path.scripts")
	info, err := os.Stat(scriptsPath + "/" + name)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "unexpected stat of script directory")
	}

	// check if expired or non-exist
	if !os.IsNotExist(err) && info.IsDir() && info.ModTime().After(updatedAt) {
		return errors.New("already up to date")
	}

	// remove expired version
	if err := os.RemoveAll(scriptsPath + "/" + name); err != nil {
		return errors.Wrap(err, "could not remove directory")
	}

	// unzip
	err = base.ScriptUser.RunWithTimeout(exec.Command(
		"unzip", scriptsPath+"/downloads/"+name+".zip", "-d", scriptsPath+"/"+name), viper.GetDuration("timeout.script.unzip"))
	if err != nil {
		return errors.Wrap(err, "could not unzip script zip file")
	}

	// set perm for compile file
	if err := base.ScriptUser.OwnRWX(scriptsPath + "/" + name + "/compile"); err != nil {
		return errors.Wrap(err, "could not set permission for script directory")
	}

	// compile
	cmd := exec.Command("bash", "./compile")
	cmd.Dir = scriptsPath + "/" + name
	err = base.ScriptUser.RunWithTimeout(cmd, viper.GetDuration("timeout.script.compile"))
	if err != nil {
		return errors.Wrap(err, "could not compile script")
	}

	return nil
}

func RunScript(name string, dir string, args ...string) error {
	path, err := filepath.Abs(viper.GetString("path.scripts") + "/" + name + "/run")
	if err != nil {
		return err
	}
	cmd := exec.Command("bash", append([]string{path}, args...)...)
	cmd.Dir = dir
	return base.ScriptUser.Run(cmd)
}

func RunScriptWithTimeout(timeout time.Duration, name string, dir string, args ...string) error {
	return base.WithTimeout(timeout, func() error {
		return RunScript(name, dir, args...)
	})
}
