package pkgconfig

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestExpand(t *testing.T) {
	vars := map[string][]byte{
		"dupa":          []byte("SPARTA"),
		"B":             []byte("XD"),
		"var":           []byte(""),
		"XD":            []byte(" "),
		"ABCDEFGHIJKLM": []byte("WAT"),
		"TABLE":         []byte("┻━┻"),
		"WAT":           []byte("ಠ_ಠ"),
	}
	cases := [...][2][]byte{{
		[]byte("THIS IS ${dupa}"),
		[]byte("THIS IS SPARTA"),
	}, {
		[]byte("${A} ${B} ${C} ${D}"),
		[]byte("${A} XD ${C} ${D}"),
	}, {
		[]byte("${var}${var}${*}${var}${*}"),
		[]byte("${*}${*}"),
	}, {
		[]byte("$$$$$$$$$$${XD}$$$$$$$$$$$$$$"),
		[]byte("$$$$$$$$$$ $$$$$$$$$$$$$$"),
	}, {
		[]byte("${ABCDEFGHIJKLM}"),
		[]byte("WAT"),
	}, {
		[]byte("(╯°□°）╯︵ ${TABLE}"),
		[]byte("(╯°□°）╯︵ ┻━┻"),
	}, {
		[]byte("${TABLE} ︵ヽ(`o´)ﾉ︵ ${TABLE}"),
		[]byte("┻━┻ ︵ヽ(`o´)ﾉ︵ ┻━┻"),
	}, {
		[]byte("x${WAT}x${TABLE}x${X}x"),
		[]byte("xಠ_ಠx┻━┻x${X}x"),
	}}
	for i, cas := range cases {
		if p := expand(cas[0], vars); !bytes.Equal(p, cas[1]) {
			t.Errorf("expected p=%q; was %q (i=%d)", cas[1], p, i)
		}
	}
}

var libgit2pc = []byte(`libdir=testdata/libgit2
includedir=testdata/include/libgit2

Name: libgit2
Description: The git library, take 2
Version: 0.20.0
URL: http://libgit2.github.com/
Libs.private: -lrt
Libs: -L${libdir} -lgit2 -Wl,-rpath -Wl,$ORIGIN
Cflags: -I${includedir}`)

var expected = &PC{
	Name:    "libgit2",
	Desc:    "The git library, take 2",
	Version: "0.20.0",
	URL:     "http://libgit2.github.com/",
	Libs: []string{
		"-Ltestdata/libgit2",
		"-lgit2",
		"-Wl,-rpath",
		"-Wl,$ORIGIN",
	},
	LibsPrivate: []string{
		"-lrt",
	},
	Cflags: []string{
		"-Itestdata/include/libgit2",
	},
	File: filepath.Join("testdata", "libgit2.pc"),
}

func TestNewPC(t *testing.T) {
	pc, err := NewPC(bytes.NewBuffer(libgit2pc))
	if err != nil {
		t.Fatalf("expected err=nil; was %q", err)
	}
	// NOTE: pc.File is set by Lookup only
	pc.File = expected.File
	if !reflect.DeepEqual(pc, expected) {
		t.Errorf("expected pc=%+v; was %+v", expected, pc)
	}
}

func TestNewPCVars(t *testing.T) {
	cases := [...]struct {
		raw  []byte
		vars map[string]string
		libs []string
	}{{
		[]byte("libdir=${libdir}/Library\n\nLibs: -L${libdir} -L${libdir}/Default"),
		map[string]string{"libdir": "/Users/rjeczalik"},
		[]string{"-L/Users/rjeczalik/Library", "-L/Users/rjeczalik/Library/Default"},
	}, {
		[]byte("libdir=${GOPATH}/lib/${GOOS}_${GOARCH}/libgit2\n\nLibs: -L${libdir}"),
		map[string]string{"GOPATH": "/home/rjeczalik", "GOOS": "linux", "GOARCH": "amd64"},
		[]string{"-L/home/rjeczalik/lib/linux_amd64/libgit2"},
	}, {
		[]byte("A=${X}${X}\nX=${A}${X}\nB=${A}${X}${A}\n\nLibs: -L${A} -L${X} -L${B}"),
		map[string]string{"X": "★"},
		[]string{"-L★★", "-L★★★", "-L★★★★★★★"},
	}}
	for i, cas := range cases {
		pc, err := NewPCVars(bytes.NewBuffer(cas.raw), cas.vars)
		if err != nil {
			t.Errorf("expected err=nil; was %q (i=%d)", err, i)
		}
		if !reflect.DeepEqual(pc.Libs, cas.libs) {
			t.Errorf("expected pc.Libs=%v; was %v (i=%d)", cas.libs, pc.Libs, i)
		}
	}
}

func TestNewPCErr(t *testing.T) {
	cases := [...][]byte{
		[]byte(""),
		[]byte(" "),
		[]byte("libdir:\n\n"),
		[]byte("libdir=/lib\nincludedir:\n\n"),
		[]byte("libdir=/lib\nincludedir=/include\n\nCflags=-I${includedir}"),
		[]byte("libdir=/lib\nincludedir=/include\n\nCflags:-I${includedir}\nName="),
	}
	for i, cas := range cases {
		if _, err := NewPC(bytes.NewBuffer(cas)); err == nil {
			t.Errorf("expected err!=nil (i=%d)", i)
		}
	}
}

func TestPCWriteTo(t *testing.T) {
	var buf bytes.Buffer
	cases := [...]struct {
		pc  *PC
		exp []byte
	}{{
		&PC{Name: "name", Libs: []string{"-A"}, Cflags: []string{"-B", "-C"}},
		[]byte("\nName: name\nLibs: -A\nCflags: -B -C\n"),
	}, {
		&PC{Name: "A", Desc: "B", Version: "C", URL: "D", Libs: []string{"-E"},
			LibsPrivate: []string{"-F"}, Cflags: []string{"-G"}, File: "I"},
		[]byte("\nName: A\nDescription: B\nVersion: C\nURL: D\nLibs.private: -F\nLibs: -E\nCflags: -G\n"),
	}, {
		&PC{Libs: []string{"-A", "", ""}, LibsPrivate: []string{"", "-B", ""},
			Cflags: []string{"", "", "-C"}},
		[]byte("\nLibs.private: -B\nLibs: -A\nCflags: -C\n"),
	}}
	for i, cas := range cases {
		buf.Reset()
		if _, err := cas.pc.WriteTo(&buf); err != nil {
			t.Errorf("expected err=nil; was %q (i=%d)", err, i)
			continue
		}
		if !bytes.Equal(buf.Bytes(), cas.exp) {
			t.Errorf("expected buf=%q; was %q (i=%d)", cas.exp, buf.Bytes(), i)
		}
	}
}

func TestPCWriteToErr(t *testing.T) {
	var buf bytes.Buffer
	cases := [...]*PC{
		{},
		{Libs: []string{""}},
		{Cflags: []string{""}},
		{Libs: []string{"-A"}},
		{Cflags: []string{"-B"}},
		{Libs: []string{""}, Cflags: []string{""}},
		{Libs: []string{""}, LibsPrivate: []string{"", ""}, Cflags: []string{"", "", ""}},
	}
	for i, cas := range cases {
		buf.Reset()
		if _, err := cas.WriteTo(&buf); err == nil {
			t.Errorf("expected err!=nil (i=%d, buf=%q)", i, buf.Bytes())
		}
	}
}

func TestLookupPC(t *testing.T) {
	if err := os.Setenv("PKG_CONFIG_PATH", "testdata"); err != nil {
		t.Fatalf("expected err=nil; was %q", err)
	}
	pc, err := LookupPC("libgit2")
	if err != nil {
		t.Fatalf("expected err=nil; was %q", err)
	}
	if !reflect.DeepEqual(pc, expected) {
		t.Errorf("expected pc=%+v; was %+v", expected, pc)
	}
}
