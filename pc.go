package pkgconfig

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// PC TODO(rjeczalik): document
type PC struct {
	Name        string
	Desc        string
	Version     string
	URL         string
	Libs        []string
	LibsPrivate []string
	Cflags      []string
	File        string
}

// ErrEmptyPC TODO(rjeczalik): document
var ErrEmptyPC = errors.New("the package configuration is empty")

// WriteTo TODO(rjeczalik): document
func (pc *PC) WriteTo(w io.Writer) (int64, error) {
	if len(pc.Libs) == 0 || len(pc.Cflags) == 0 {
		return 0, ErrEmptyPC
	}
	var buf bytes.Buffer
	buf.WriteByte('\n')
	// TODO(rjeczalik): map interation order?
	for _, item := range []struct{ s, v string }{
		{"Name", pc.Name},
		{"Description", pc.Desc},
		{"Version", pc.Version},
		{"URL", pc.URL},
		{"Libs.private", strings.TrimSpace(strings.Join(pc.LibsPrivate, " "))},
		{"Libs", strings.TrimSpace(strings.Join(pc.Libs, " "))},
		{"Cflags", strings.TrimSpace(strings.Join(pc.Cflags, " "))},
	} {
		if item.v != "" {
			buf.WriteString(item.s)
			buf.WriteByte(':')
			buf.WriteByte(' ')
			buf.WriteString(item.v)
			buf.WriteByte('\n')
		}
	}
	if buf.Len() == 1 {
		return 0, ErrEmptyPC
	}
	return io.Copy(w, &buf)
}

func max(n, m int) int {
	if n > m {
		return n
	}
	return m
}

func flatsplit(s, sep string) []string {
	v := strings.Split(s, sep)
	for i := 0; i < len(v); i++ {
		if v[i] == "" {
			v = append(v[:i], v[i+1:]...)
			i -= 1
		}
	}
	return v
}

func expand(p []byte, vars map[string][]byte) []byte {
	for n, m := bytes.IndexByte(p, '$'), 0; n != -1; n = bytes.IndexByte(p[m:], '$') {
		m += n
		if m+3 < len(p) && p[m+1] == '{' {
			if l := bytes.IndexByte(p[m+2:], '}'); l != -1 {
				if value, ok := vars[string(p[m+2:m+l+2])]; ok {
					p = append(p[:m], append(value, p[m+l+3:]...)...)
					m -= l + 3 - len(value)
				}
			}
		}
		m = max(0, m+1)
	}
	return p
}

type state struct {
	n string
	c byte
}

var (
	parseVar = &state{"variable", '='}
	parseKey = &state{"keyword", ':'}
)

// NewPC TODO(rjeczalik): document
func NewPC(r io.Reader) (*PC, error) {
	return NewPCVars(r, make(map[string]string))
}

// NewPCVars TODO(rjeczalik): document
func NewPCVars(r io.Reader, vars map[string]string) (pc *PC, err error) {
	pc = &PC{}
	var (
		buf = bufio.NewReader(r)
		m   = make(map[string][]byte, len(vars))
		st  = parseVar
		p   []byte
		c   int
	)
	fail := func() error {
		return fmt.Errorf("malformed %s line: %q", st.n, p)
	}
	for n, v := range vars {
		m[n] = []byte(v)
	}
	for {
		p, err = buf.ReadBytes('\n')
		p = bytes.TrimSpace(p)
		if len(p) == 0 {
			if err == io.EOF {
				err = nil
				if c == 0 {
					err = io.ErrUnexpectedEOF
				}
				return
			}
			if err != nil {
				return
			}
			st = parseKey
			continue
		}
		c += len(p)
		n := bytes.IndexByte(p, st.c)
		if n == -1 {
			err = fail()
			return
		}
		name := string(bytes.TrimSpace(p[:n]))
		if len(name) == 0 {
			err = fail()
			return
		}
		value := bytes.TrimSpace(p[n+1:])
		switch st {
		case parseVar:
			value = expand(value, m)
			m[name] = value
		case parseKey:
			v := string(expand(value, m))
			switch strings.ToLower(name) {
			case "name":
				pc.Name = v
			case "description":
				pc.Desc = v
			case "version":
				pc.Version = v
			case "url":
				pc.URL = v
			case "libs":
				// BUG(rjeczalik): Handle spaces in paths.
				pc.Libs = flatsplit(v, " ")
			case "libs.private":
				// BUG(rjeczalik): Handle spaces in paths.
				pc.LibsPrivate = flatsplit(v, " ")
			case "cflags":
				// BUG(rjeczalik): Handle spaces in paths.
				pc.Cflags = flatsplit(v, " ")
			}
		}
	}
}

var defaultPaths []string

// LookupPC TODO(rjeczalik): document
func LookupPC(pkg string) (*PC, error) {
	var (
		paths = strings.Split(os.Getenv("PKG_CONFIG_PATH"), string(os.PathListSeparator))
		err   error
		pc    *PC
		f     *os.File
	)
	paths = append(paths, defaultPaths...)
	for _, path := range paths {
		file := filepath.Join(path, pkg+".pc")
		if f, err = os.Open(file); err == nil {
			pc, err = NewPC(f)
			f.Close()
			if err == nil {
				pc.File = file
				return pc, nil
			}
		}
	}
	return nil, err
}
