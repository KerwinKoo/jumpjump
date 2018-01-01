package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

// ReadFile read buf from path and return string buf
// return string
func ReadFile(path string) (string, error) {
	fi, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	return string(fd), nil
}

// ReadFileByte read buf from path and return string buf
// return byte
func ReadFileByte(path string) ([]byte, error) {
	fi, err := os.Open(path)
	if err != nil {
		return []byte(""), err
	}
	defer fi.Close()
	fd, err := ioutil.ReadAll(fi)
	return fd, nil
}

// PathExist check file exist or not
func PathExist(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

// WriteFileWithLock write buf to file with LINUX LOCK
func WriteFileWithLock(path string, buf []byte) error {
	if !PathExist(path) {
		_, err := os.Create(path)
		if err != nil {
			return err
		}
	}

	fwriter, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer fwriter.Close()

	syscall.Flock(int(fwriter.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(fwriter.Fd()), syscall.LOCK_UN)

	_, err = fwriter.Write(buf)
	if err != nil {
		return err
	}

	return nil
}

// WriteAppendWithLock write content append filename with file mutx lock
func WriteAppendWithLock(path string, buf []byte) error {
	if !PathExist(path) {
		_, err := os.Create(path)
		if err != nil {
			return err
		}
	}
	fwriter, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer fwriter.Close()

	syscall.Flock(int(fwriter.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(fwriter.Fd()), syscall.LOCK_UN)
	_, err = fwriter.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

// ReadFileWithLock read file with LINUX LOCK
func ReadFileWithLock(path string) ([]byte, error) {
	fi, err := os.Open(path)
	if err != nil {
		return []byte(""), err
	}
	defer fi.Close()

	syscall.Flock(int(fi.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(fi.Fd()), syscall.LOCK_UN)
	fd, err := ioutil.ReadAll(fi)
	return fd, err
}

// Md5Checksum2File write the sha256sums of checkTargetFilename file
// into md5Filename
// checkTargetFilename: file need to be md5 checksum
func Md5Checksum2File(checkTargetFilename, md5Filename string) {
	file, err := os.Open(checkTargetFilename)
	if err != nil {
		log.Printf("Error on opening file: %s\n", checkTargetFilename)
		return
	}

	md5h := md5.New()
	io.Copy(md5h, file)

	//filename 	md5checksum
	md5Result := fmt.Sprintf("%s\t%x\n", checkTargetFilename, md5h.Sum([]byte("")))

	md5File, err := os.OpenFile(md5Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		log.Printf("Error on opening file: %s\n", md5Filename)
		return
	}
	defer md5File.Close()

	// Acquire the lock
	syscall.Flock(int(md5File.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(md5File.Fd()), syscall.LOCK_UN)
	_, err = md5File.Write([]byte(md5Result)) //md5
	if err != nil {
		log.Printf("Error on writting file: %s\n", md5Filename)
		return
	}
}

// ListFiles return the file list of th path
// the result will egnore directories, only files, and only filename
func ListFiles(path string) ([]string, error) {
	var flist []string

	err := filepath.Walk(path, func(filepath string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		isDir := f.IsDir()
		if isDir == false {
			flist = append(flist, f.Name())
			return nil
		}

		return nil
	})

	if err != nil {
		return flist, err
	}

	return flist, nil
}

// ListFilesAbsPath list the files under path
// the result will return the file name include ABS name
func ListFilesAbsPath(path string) ([]string, error) {
	var absflist []string

	err := filepath.Walk(path, func(filepath string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		isDir := f.IsDir()
		if isDir == false {
			log.Println("file:", filepath)
			absflist = append(absflist, filepath)
			return nil
		}

		return nil
	})

	if err != nil {
		return absflist, err
	}

	return absflist, nil
}
