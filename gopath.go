package pkgconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var defaultGopath = gopath(os.PathListSeparator)

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

func gopath(sep rune) []string {
	var paths = strings.Split(os.Getenv("GOPATH"), string(sep))
	if wd != "" {
		if i := strings.Index(wd, src[sep]); i != -1 {
			tmp := make([]string, 0, len(paths)+1)
			tmp = append(tmp, wd[:i])
			paths = append(tmp, paths...)
		}
	}
	for i := 1; i < len(paths); i++ {
		if paths[i] == paths[0] {
			paths = append(paths[:i], paths[i+1:]...)
			break
		}
	}
	return paths
}

// GopathLibrary TODO(rjeczalik): document
func GopathLibrary(path, pkg string) (include, lib string) {
	include = filepath.Join(path, "include", pkg)
	lib = filepath.Join(path, "lib", runtime.GOOS+"_"+runtime.GOARCH, pkg)
	return
}

func walkgopath(pkg string, fn func(string, string, string) bool) bool {
	for _, path := range defaultGopath {
		include, lib := GopathLibrary(path, pkg)
		if existDir(include, lib) != nil {
			continue
		}
		if !fn(path, include, lib) {
			return true
		}
	}
	return false
}

// LookupGopath TODO(rjeczalik): document
func LookupGopath(pkg string) (*PC, error) {
	var (
		vars = map[string]string{"GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH}
		err  error
		pc   *PC
		f    *os.File
	)
	look := func(path, _, lib string) bool {
		file := filepath.Join(lib, pkg+".pc")
		vars["GOPATH"] = filepath.ToSlash(path)
		if f, err = os.Open(file); err == nil {
			pc, err = NewPCVars(f, vars)
			f.Close()
			if err == nil {
				pc.File = file
				return false
			}
		}
		return true
	}
	if !walkgopath(pkg, look) {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("no library found in $GOPATH: " + pkg)
	}
	return pc, nil
}

// GenerateGopath TODO(rjeczalik): document
func GenerateGopath(pkg string) (*PC, error) {
	var pc *PC
	gen := func(path, include, lib string) bool {
		pc = &PC{
			Libs: []string{
				"-L" + lib,
				"-l" + strings.TrimLeft(pkg, "lib"),
				"-Wl,-rpath", "-Wl,$ORIGIN",
			},
			Cflags: []string{"-I" + include},
		}
		return false
	}
	if !walkgopath(pkg, gen) {
		return nil, errors.New("no library found in $GOPATH: " + pkg)
	}
	return pc, nil
}
