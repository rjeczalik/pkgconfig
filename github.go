package pkgconfig

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var githubProj string

func init() {
	if wd, err := os.Getwd(); err == nil {
		githubProj = extractproj(wd, os.PathSeparator)
	}
}

var src = map[rune]string{'/': "/src/", '\\': `\src\`}

var projpath = map[rune]func(string) string{
	'/':  func(s string) string { return s },
	'\\': func(s string) string { return strings.Replace(s, "\\", "/", -1) },
}

func extractproj(path string, sep rune) (proj string) {
	n := strings.LastIndex(path, src[sep])
	if n == -1 {
		return
	}
	n += len(src[sep])
	m := n
	for i := 0; i < 3; i++ {
		l := strings.Index(path[m:], string(sep))
		if l == -1 {
			m = len(path) + 1
			break
		}
		m += l + 1
	}
	proj = projpath[sep](string(path[n : m-1]))
	if !strings.HasPrefix(proj, "github.com") {
		proj = ""
	}
	return
}

// LookupGithub TODO(rjeczalik): document
func LookupGithub(pkg string) (*PC, error) {
	if githubProj == "" {
		return nil, errors.New(`unable to guess project's URL from $CWD`)
	}
	return LookupGithubProj(pkg, githubProj)
}

var targets = map[string]struct{}{
	"darwin_386":    {},
	"darwin_amd64":  {},
	"freebsd_386":   {},
	"freebsd_amd64": {},
	"freebsd_arm":   {},
	"linux_386":     {},
	"linux_amd64":   {},
	"linux_arm":     {},
	"netbsd_386":    {},
	"netbsd_amd64":  {},
	"netbsd_arm":    {},
	"openbsd_386":   {},
	"openbsd_amd64": {},
	"plan9_386":     {},
	"plan9_amd64":   {},
	"windows_386":   {},
	"windows_amd64": {},
}

const targetsMinKeyLen = 9

func validFile(path, pkg string) bool {
	switch {
	case strings.HasPrefix(path, "include/"):
		i := len("include/")
		if len(path) < i+len(pkg)+1 {
			return false
		}
		n := strings.Index(path[i:], "/")
		if n == -1 {
			return false
		}
		return string(path[i:i+n]) == pkg
	case strings.HasPrefix(path, "lib/"):
		i := len("lib/")
		if len(path) < i+targetsMinKeyLen+1+len(pkg)+1 {
			return false
		}
		n := strings.Index(path[i:], "/")
		if n == -1 {
			return false
		}
		if _, ok := targets[string(path[i:i+n])]; !ok {
			return false
		}
		i += n + 1
		if n = strings.Index(path[i:], "/"); n == -1 {
			return false
		}
		return string(path[i:i+n]) == pkg
	}
	return false
}

var replacefile = map[rune]func(string) string{
	'/':  func(s string) string { return s },
	'\\': func(s string) string { return strings.Replace(s, "/", "\\", -1) },
}

func copyFile(gopath string, f *zip.File) (err error) {
	rc, err := f.Open()
	if err != nil {
		return
	}
	defer rc.Close()
	dir := filepath.Join(gopath, replacefile[os.PathSeparator](path.Dir(f.Name)))
	if err = os.MkdirAll(dir, 0755); err != nil {
		return
	}
	file, err := os.OpenFile(filepath.Join(dir, path.Base(f.Name)), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	_, err = io.Copy(file, rc)
	file.Close()
	return
}

// LookupGithunProj TODO(rjeczalik): document
func LookupGithubProj(pkg, proj string) (*PC, error) {
	path := os.Getenv("GOPATH")
	if p := strings.Split(path, string(os.PathListSeparator)); len(path) != 0 {
		path = p[0]
	}
	if path == "" {
		return nil, errors.New("$GOPATH is empty")
	}
	url := fmt.Sprintf("http://%s/releases/download/pkg-config/%s.zip", proj, pkg)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return nil, errors.New("not found: " + url)
	}
	f, err := ioutil.TempFile("", pkg)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(f, res.Body); err != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, err
	}
	f.Close()
	r, err := zip.OpenReader(f.Name())
	if err != nil {
		return nil, err
	}
	defer func() {
		r.Close()
		os.Remove(f.Name())
	}()
	for _, f := range r.File {
		// Filter out directories.
		if f.Name[len(f.Name)-1] != '/' {
			if !validFile(f.Name, pkg) {
				return nil, fmt.Errorf("unexcpected file %q", f.Name)
			}
			if err = copyFile(path, f); err != nil {
				return nil, err
			}
		}
	}
	return LookupGopath(pkg)
}
