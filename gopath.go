package pkgconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func existDir(dirs ...string) (err error) {
	var fi os.FileInfo
	for _, dir := range dirs {
		if fi, err = os.Stat(dir); err != nil {
			return
		}
		if !fi.IsDir() {
			err = fmt.Errorf("%s is not a directory", dir)
			return
		}
	}
	return
}

func existFile(files ...string) (err error) {
	var fi os.FileInfo
	for _, file := range files {
		if fi, err = os.Stat(file); err != nil {
			return
		}
		if fi.IsDir() {
			err = fmt.Errorf("%s is not a file", file)
			return
		}
	}
	return
}

func Gopath(path, pkg string) (include, lib string) {
	include = filepath.Join(path, "include", pkg)
	lib = filepath.Join(path, "lib", runtime.GOOS+"_"+runtime.GOARCH, pkg)
	return
}

// LookupGopath TODO(rjeczalik): document
func LookupGopath(pkg string) (*PC, error) {
	var (
		paths = strings.Split(os.Getenv("GOPATH"), string(os.PathListSeparator))
		vars  = map[string]string{"GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH}
		err   error
		pc    *PC
		f     *os.File
	)
LOOP:
	for _, path := range paths {
		include, lib := Gopath(path, pkg)
		if err = existDir(include, lib); err != nil {
			continue LOOP
		}
		file := filepath.Join(lib, pkg+".pc")
		vars["GOPATH"] = path
		if f, err = os.Open(file); err == nil {
			pc, err = NewPCVars(f, vars)
			f.Close()
			if err == nil {
				pc.File = file
				return pc, nil
			}
		}
	}
	return nil, err
}
