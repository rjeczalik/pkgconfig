package pkgconfig

import (
	"bytes"
	"errors"
	"os"
	"reflect"
	"testing"
)

func TestNewPkgArgs(t *testing.T) {
	cases := [...]struct {
		args []string
		exp  *Pkg
	}{{
		[]string{"lib"},
		&Pkg{Packages: []string{"lib"}},
	}, {
		[]string{"lib1", "lib2", "lib3"},
		&Pkg{Packages: []string{"lib1", "lib2", "lib3"}},
	}, {
		[]string{"--cflags", "lib1", "lib2", "lib3"},
		&Pkg{Cflags: true, Packages: []string{"lib1", "lib2", "lib3"}},
	}, {
		[]string{"--libs", "lib1", "lib2", "lib3"},
		&Pkg{Libs: true, Packages: []string{"lib1", "lib2", "lib3"}},
	}, {
		[]string{"--cflags", "--libs", "lib1", "lib2", "lib3"},
		&Pkg{Cflags: true, Libs: true, Packages: []string{"lib1", "lib2", "lib3"}},
	}, {
		[]string{"--cflags", "--libs", "--libs.private", "-XD", "lib1", "lib2", "lib3"},
		&Pkg{Cflags: true, Libs: true, Packages: []string{"lib1", "lib2", "lib3"}},
	}}
	for i, cas := range cases {
		if pkg := NewPkgArgs(cas.args); pkg == nil || !reflect.DeepEqual(pkg, cas.exp) {
			t.Errorf("expected pkg=%+v; was %+v (i=%d)", cas.exp, pkg, i)
		}
	}
}

func TestPkgResolve(t *testing.T) {
	all := map[string]*PC{
		"A": &PC{File: "A.pc"},
		"B": &PC{File: "B.pc"},
		"C": &PC{File: "C.pc"},
		"D": &PC{File: "D.pc"},
		"E": &PC{File: "E.pc"},
	}
	lu := func(pkg string) (*PC, error) {
		pc, ok := all[pkg]
		if !ok {
			return nil, errors.New("not found")
		}
		return pc, nil
	}
	cases := []struct {
		pkgs []string
		pc   map[string]*PC
	}{{
		[]string{"A"},
		map[string]*PC{"A": all["A"]},
	}, {
		[]string{"A", "A", "A"},
		map[string]*PC{"A": all["A"]},
	}, {
		[]string{"A", "B", "C"},
		map[string]*PC{"B": all["B"], "C": all["C"], "A": all["A"]},
	}, {
		[]string{"A", "B", "C", "D", "E"},
		map[string]*PC{"A": all["A"], "B": all["B"], "C": all["C"], "D": all["D"], "E": all["E"]},
	}, {
		[]string{"A", "A", "A", "B", "B", "B", "E", "E", "E"},
		map[string]*PC{"A": all["A"], "E": all["E"], "B": all["B"]},
	}}
	for i, cas := range cases {
		pkg := &Pkg{
			Packages: cas.pkgs,
			Lookup:   lu,
		}
		if err := pkg.Resolve(); err != nil {
			t.Errorf("expected err=nil; was %q (i=%d)", err, i)
			continue
		}
		if !reflect.DeepEqual(pkg.pc, cas.pc) {
			t.Errorf("expected pkg.pc=%+v; was %+v (i=%d)", cas.pc, pkg.pc, i)
		}
	}
	casesErr := [][]string{
		{},
		{"F"},
		{"A", "B", "C", "D", "F"},
	}
	for i, cas := range casesErr {
		pkg := &Pkg{
			Packages: cas,
			Lookup:   lu,
		}
		if err := pkg.Resolve(); err == nil {
			t.Errorf("expected err==nil (i=%d)", i)
		}
		if len(pkg.pc) != 0 {
			t.Errorf("expected len(pkg.pc)==0; was %d (i=%d)", len(pkg.pc), i)
		}
	}
}

func TestPkgResolveAccept(t *testing.T) {
	for i, env := range []string{"GOPATH", "PKG_CONFIG_PATH"} {
		if err := os.Setenv(env, "testdata"); err != nil {
			t.Errorf("expected err=nil; was %q (i=%d)", err, i)
		}
		pkg := &Pkg{Packages: []string{"libgit2"}}
		if err := pkg.Resolve(); err != nil {
			t.Errorf("expected err=nil; was %q (i=%d)", err, i)
		}
		if err := os.Setenv(env, ""); err != nil {
			t.Errorf("expected err=nil; was %q (i=%d)", err, i)
		}
	}
}

func TestPkgWriteTo(t *testing.T) {
	var buf bytes.Buffer
	newpkg := func(cflags, libs bool, c, l map[string][]string) *Pkg {
		pkg := &Pkg{Cflags: cflags, Libs: libs, pc: make(map[string]*PC, len(c))}
		for k := range c {
			pkg.pc[k] = &PC{Cflags: c[k], Libs: l[k]}
		}
		return pkg
	}
	cases := []struct {
		cflags bool
		c      map[string][]string
		libs   bool
		l      map[string][]string
		exp    []byte
	}{{
		true, map[string][]string{
			"A": {"-ca"},
			"B": {"-cb"},
		},
		true, map[string][]string{
			"A": {"-la"},
			"B": {"-lb"},
		},
		[]byte("-ca -cb -la -lb\n"),
	}, {
		true, map[string][]string{
			"A": {"-ca", "-a", "-a"},
			"B": {"-b", "-cb", "-b"},
		},
		true, map[string][]string{
			"A": {"-a", "-la"},
			"B": {"-lb", "-b"},
		},
		[]byte("-ca -a -b -cb -la -lb\n"),
	}, {
		true, map[string][]string{
			"A": {"-ca", "-l", "-a", "-a"},
			"B": {"-b", "-l", "-cb", "-b", "-x"},
		},
		true, map[string][]string{
			"A": {"-a", "-l", "-la"},
			"B": {"-lb", "-l", "-b", "-x"},
		},
		[]byte("-ca -l -a -b -cb -x -la -lb\n"),
	}, {
		true, map[string][]string{
			"A": {"-ca", "-l", "-a", "-a"},
			"B": {"-b", "-l", "-cb", "-b", "-x"},
		},
		false, map[string][]string{
			"A": {"-a", "-l", "-la"},
			"B": {"-lb", "-l", "-b", "-x"},
		},
		[]byte("-ca -l -a -b -cb -x\n"),
	}, {
		false, map[string][]string{
			"A": {"-ca", "-l", "-a", "-a"},
			"B": {"-b", "-l", "-cb", "-b", "-x"},
		},
		true, map[string][]string{
			"A": {"-a", "-l", "-la"},
			"B": {"-lb", "-l", "-b", "-x"},
		},
		[]byte("-a -l -la -lb -b -x\n"),
	}}
	for i, cas := range cases {
		buf.Reset()
		pkg := newpkg(cas.cflags, cas.libs, cas.c, cas.l)
		if _, err := pkg.WriteTo(&buf); err != nil {
			t.Errorf("expected err=nil; was %q (i=%d)", err, i)
			continue
		}
		if !bytes.Equal(buf.Bytes(), cas.exp) {
			t.Errorf("expected buf=%q; was %q (i=%d)", cas.exp, buf.Bytes(), i)
		}
	}
}
