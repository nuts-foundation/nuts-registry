package test

import (
	io2 "github.com/nuts-foundation/nuts-go-test/io"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

type TestRepo struct {
	Directory string
}

func NewTestRepo(t *testing.T) (*TestRepo, error) {
	return NewTestRepoFrom(t, "")
}

func NewTestRepoFrom(t *testing.T, sourceDir string) (*TestRepo, error) {
	dir := io2.TestDirectory(t)
	if sourceDir != "" {
		err := copyDir(sourceDir, dir)
		if err != nil {
			return nil, err
		}
	}
	return &TestRepo{Directory: dir}, nil
}

func (r TestRepo) Cleanup() {
	os.RemoveAll(r.Directory)
}

func (r TestRepo) ImportDir(sourceDirectory string) error {
	return copyDir(sourceDirectory, r.Directory)
}

func (r TestRepo) ImportFileAs(sourceFile string, targetFile string) error {
	absTargetFile := filepath.Join(r.Directory, targetFile)
	err := os.MkdirAll(filepath.Dir(absTargetFile), os.ModePerm)
	if err != nil {
		return err
	}
	return copyFile(sourceFile, absTargetFile)
}

func copyDir(src string, dst string) error {
	dir, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dst, os.ModePerm); err != nil {
		return err
	}
	for _, entry := range dir {
		sourceFile := filepath.Join(src, entry.Name())
		targetFile := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			err := copyDir(sourceFile, targetFile)
			if err != nil {
				return err
			}
			continue
		}
		if err := copyFile(sourceFile, targetFile); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)

	return err
}
