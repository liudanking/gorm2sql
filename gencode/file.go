package gencode

import (
	"os"
	"os/user"
	"path/filepath"
)

func WriteFile(fn string, s string) error {
	f, err := os.OpenFile(fn, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(s))
	return err
}

func AbsPath(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	if path[0] == '~' {
		user, err := user.Current()
		if err != nil {
			return "", err
		}
		path = filepath.Join(user.HomeDir, path[1:])
	}

	return filepath.Abs(path)
}
