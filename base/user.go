package base

import (
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"syscall"
	"time"
)

type User struct {
	User *user.User
	Uid  uint32
	Gid  uint32
}

var (
	CompileUser User
	RunUser     User
)

func (u *User) Init(username string) error {
	var err error
	if u.User, err = user.Lookup(username); err != nil {
		return err
	}
	gid, err := strconv.ParseUint(u.User.Gid, 10, 32)
	if err != nil {
		return err
	}
	u.Gid = uint32(gid)
	uid, err := strconv.ParseUint(u.User.Uid, 10, 32)
	if err != nil {
		return err
	}
	u.Uid = uint32(uid)
	return nil
}

func (u *User) Run(cmd *exec.Cmd) error {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: u.Uid, Gid: u.Gid}
	return cmd.Run()
}

func (u *User) RunWithTimeout(cmd *exec.Cmd, timeout time.Duration) error {
	return WithTimeout(timeout, func() error {
		return u.Run(cmd)
	})
}

func (u *User) OwnRWX(path string) error {
	if err := os.Chmod(path, 0700); err != nil {
		return err
	}
	return os.Chown(path, int(u.Uid), int(u.Gid))
}
