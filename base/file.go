package base

import "os"

func RemoveBuffer() error {
	err := os.RemoveAll(ScriptPath)
	if err != nil {
		return err
	}
	err = os.RemoveAll(RunPath)
	if err != nil {
		return err
	}
	return nil
}
