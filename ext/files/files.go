package files

import (
	"os"
	"path/filepath"
)

// Exist returns a boolean indicating whether a file or directory exists.
func Exist(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// NotExist returns a boolean indicating whether a file or directory does not exist.
func NotExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return os.IsNotExist(err)
}

// IsDir returns a boolean indicating whether a path is directory.
func IsDir(filePath string) bool {
	fi, err := os.Stat(filePath)
	return err == nil && fi.IsDir()
}

// Open opens the named file with flag os.O_WRONLY|os.O_APPEND|os.O_CREATE and perm 0666.
// If the file is not exists, it will be created automatically.
func Open(filename string) (*os.File, error) {
	return OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
}

// OpenFile opens the named file with specified flag (O_RDONLY etc.) and perm (0666 etc.) if applicable.
// If the file is not exists, it will be created automatically.
// If successful, methods on the returned File can be used for I/O.
// If there is an error, it will be of type *PathError.
func OpenFile(filename string, flag int, perm os.FileMode) (*os.File, error) {
	dir := filepath.Dir(filename)
	if NotExist(dir) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}
	return os.OpenFile(filename, flag, perm)
}
