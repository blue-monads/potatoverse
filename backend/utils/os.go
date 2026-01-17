package xutils

import (
	"archive/zip"
	"errors"
	"io"
	"net"
	"os"
	"path"
	"time"

	"github.com/blue-monads/potatoverse/backend/utils/qq"
)

func FileExists(dpath, file string) bool {
	_, err := os.Stat(path.Join(dpath, file))
	if err == nil {
		return true
	}
	return !errors.Is(err, os.ErrNotExist)
}

func CreateIfNotExits(fpath string) error {
	if _, err := os.Stat(fpath); errors.Is(err, os.ErrNotExist) {
		return os.Mkdir(fpath, os.ModePerm)
	} else {
		if err != nil {
			return err
		}
	}

	return nil
}

func Die(args ...any) {

	qq.Println(args...)
	os.Exit(1)

}

func Copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func ExtractZip(zfile, ofolder string) error {
	z, err := zip.OpenReader(zfile)
	if err != nil {
		panic(err)
	}
	defer z.Close()

	for _, file := range z.File {
		fileReader, err := file.Open()
		if err != nil {
			return err
		}

		defer fileReader.Close()

		extractedFilePath := ofolder + string(os.PathSeparator) + file.Name
		extractedFile, err := os.Create(extractedFilePath)
		if err != nil {
			return err
		}

		defer extractedFile.Close()

		_, err = io.Copy(extractedFile, fileReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func CollapseTimestampId() int64 {
	interval := int64(time.Minute * 15) // 15 minutes

	now := time.Now().Unix()

	rounded := now - (now % interval)

	return rounded
}
