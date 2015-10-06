package xattr

import (
	xa "github.com/davecheney/xattr"
	"strings"
)

func GetXattr(path, name string) ([]byte, bool, error) {
	_ = "breakpoint"
	var attr []byte
	attr, err := xa.Getxattr(path, name)
	if err != nil && !strings.Contains(err.Error(), "errno 0") {
		if strings.Contains(err.Error(), "not found") {
			return []byte(""), false, nil
		}
		return attr, false, err
	}
	return attr, true, nil
}

func SetXattr(path, name string, val []byte) error {
	_ = "breakpoint"
	err := xa.Setxattr(path, name, val)
	if err != nil && !strings.Contains(err.Error(), "errno 0") {
		return err
	}
	return nil
}
