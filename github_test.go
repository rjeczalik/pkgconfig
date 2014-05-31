package pkgconfig

import "testing"

func TestExtractProj(t *testing.T) {
	cases := []struct {
		path map[rune]string
		exp  string
	}{{
		map[rune]string{
			'/':  "/home/rjeczalik/workspace/src/github.com/libgit2/git2go",
			'\\': `C:\Users\rjeczalik\My Workspace\src\github.com\libgit2\git2go`,
		},
		"github.com/libgit2/git2go",
	}, {
		map[rune]string{
			'/':  "/Users/rjeczalik/src/github.com/rjeczalik/fakerpc/cmd/fakerpc",
			'\\': `C:\Documents and Settings\src\github.com\rjeczalik\fakerpc\cmd\fakerpc`,
		},
		"github.com/rjeczalik/fakerpc",
	}, {
		map[rune]string{
			'/':  "~/src/github.com/rjeczalik/crontz/cmd/crontz",
			'\\': `%userprofile%\src\github.com\rjeczalik\crontz\cmd\crontz`,
		},
		"github.com/rjeczalik/crontz",
	}, {
		map[rune]string{
			'/':  "~/src/github.com/gocircuit/circuit/src/github.com/codegangsta/cli",
			'\\': `%userprofile%\src\github.com\gocircuit\circuit\src\github.com\codegangsta\cli`,
		},
		"github.com/codegangsta/cli",
	}, {
		map[rune]string{
			'/':  "/Users/rjeczalik/src/bitbucket.org/rjeczalik/benchnosql",
			'\\': `C:\Users\rjeczalik\src\bitbucker.org\rjeczalik\benchnosql`,
		},
		"",
	}, {
		map[rune]string{
			'/':  "./src/codeplex.com/rjeczalik/casablanca",
			'\\': `.\src\codeplex.com\rjeczalik\casablanca`,
		},
		"",
	}}
	for i, cas := range cases {
		for _, sep := range []rune{'/', '\\'} {
			if s := extractproj(cas.path[sep], sep); s != cas.exp {
				t.Errorf("expected s=%q; was %q (sep=%q, i=%d)", cas.exp, s, sep, i)
			}
		}
	}
}

func TestValidFile(t *testing.T) {
	valid := []string{
		"include/libgit2/git2/sys/index.h",
		"include/libgit2/git2/sys/refdb_backend.h",
		"include/libgit2/git2/attr.h",
		"include/libgit2/git2.h",
		"lib/windows_amd64/libgit2/libgit2.pc",
		"lib/windows_amd64/libgit2/libgit2.dll",
		"lib/windows_amd64/libgit2/libgit2.dll.a",
		"lib/linux_amd64/libgit2/libgit2.pc",
		"lib/linux_amd64/libgit2/libgit2.so",
		"lib/linux_386/libgit2/libgit2.pc",
		"lib/linux_386/libgit2/libgit2.so",
		"lib/darwin_386/libgit2/libgit2.pc",
		"lib/darwin_386/libgit2/libgit2.dylib",
		"lib/darwin_amd64/libgit2/libgit2.pc",
		"lib/darwin_amd64/libgit2/libgit2.dylib",
		"lib/windows_386/libgit2/libgit2.pc",
		"lib/windows_386/libgit2/libgit2.dll",
		"lib/windows_386/libgit2/libgit2.dll.a",
	}
	for _, file := range valid {
		if !validFile(file, "libgit2") {
			t.Errorf("expected file=%q to be valid", file)
		}
	}
	invalid := []string{
		"",
		"/",
		"foo",
		"foo/",
		"foo/windows_amd64/bar",
		"foo/windows_amd64/bar/",
		"foo/windows_amd64/bar/libbar.so",
		"include/bar",
		"include/bar/",
		"lib/windows_amd64/bar",
		"lib/darwin_amd64/bar/",
		"lib/windar_amd64/bar/libbar.so",
		"lib/windows_arm/bar/libbar.so",
		"include/libgit2",
		"include/libgit2/",
		"include/libgit2/git2/sys/index.h",
		"include/libgit2/git2.h",
		"lib",
		"lib/",
		"lib/darwin_386",
		"lib/darwin_386/",
		"lib/windows_386/libgit2",
		"lib/windows_386/libgit2/",
		"lib/windows_386/libgit2/libgit2.pc",
		"lib/windows_amd64/libgit2",
		"lib/windows_amd64/libgit2/",
		"lib/windows_amd64/libgit2/libgit2.dll",
	}
	for _, file := range invalid {
		if validFile(file, "foo") {
			t.Errorf("expected path=%q to be invalid", file)
		}
	}
}

func TestLookupGithub(t *testing.T) {

}
