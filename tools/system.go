package tools

import (
	"os"
	"path/filepath"
	"strings"
)

// GetCurrentDirectory
func GetCurrentDirectory() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0])) //返回绝对路径  filepath.Dir(os.Args[0])去除最后一个元素的路径
	if err != nil {
		return "", err
	}
	return strings.Replace(dir, "\\", "/", -1), nil //将\替换成/
}
