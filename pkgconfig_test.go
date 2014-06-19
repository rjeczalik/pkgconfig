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
		pc   []*PC
	}{{
		[]string{"A"},
		[]*PC{all["A"]},
	}, {
		[]string{"A", "A", "A"},
		[]*PC{all["A"]},
	}, {
		[]string{"A", "B", "C"},
		[]*PC{all["A"], all["B"], all["C"]},
	}, {
		[]string{"A", "B", "C", "D", "E"},
		[]*PC{all["A"], all["B"], all["C"], all["D"], all["E"]},
	}, {
		[]string{"A", "A", "A", "B", "B", "B", "E", "E", "E"},
		[]*PC{all["A"], all["B"], all["E"]},
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
	newpkg := func(cflags, libs bool, c, l [][]string) *Pkg {
		pkg := &Pkg{Cflags: cflags, Libs: libs, pc: make([]*PC, len(c))}
		for i := range c {
			pkg.pc[i] = &PC{Cflags: c[i], Libs: l[i]}
		}
		return pkg
	}
	cases := []struct {
		cflags bool
		c      [][]string
		libs   bool
		l      [][]string
		exp    []byte
	}{{
		true, [][]string{
			0: {"-ca"},
			1: {"-cb"},
		},
		true, [][]string{
			0: {"-la"},
			1: {"-lb"},
		},
		[]byte("-ca -cb -la -lb\n"),
	}, {
		true, [][]string{
			0: {"-ca", "-a", "-a"},
			1: {"-b", "-cb", "-b"},
		},
		true, [][]string{
			0: {"-a", "-la"},
			1: {"-lb", "-b"},
		},
		[]byte("-ca -a -b -cb -la -lb\n"),
	}, {
		true, [][]string{
			0: {"-ca", "-l", "-a", "-a"},
			1: {"-b", "-l", "-cb", "-b", "-x"},
		},
		true, [][]string{
			0: {"-a", "-l", "-la"},
			1: {"-lb", "-l", "-b", "-x"},
		},
		[]byte("-ca -l -a -b -cb -x -la -lb\n"),
	}, {
		true, [][]string{
			0: {"-ca", "-l", "-a", "-a"},
			1: {"-b", "-l", "-cb", "-b", "-x"},
		},
		false, [][]string{
			0: {"-a", "-l", "-la"},
			1: {"-lb", "-l", "-b", "-x"},
		},
		[]byte("-ca -l -a -b -cb -x\n"),
	}, {
		false, [][]string{
			0: {"-ca", "-l", "-a", "-a"},
			1: {"-b", "-l", "-cb", "-b", "-x"},
		},
		true, [][]string{
			0: {"-a", "-l", "-la"},
			1: {"-lb", "-l", "-b", "-x"},
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
