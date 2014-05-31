// cmd/pkg-config is a Go-centric and GOPATH-aware pkg-config replacement
// for use with the cgo tool.
//
// Using cmd/pkg-config with the cgo command
//
// To use cmd/pkg-config go install it and ensure it's in the PATH. A NOTE for
// Linux users: the cmd/pkg-config must be present before original
// /usr/bin/pkg-config in the PATH list; it is not advised to replace it,
// as cmd/pkg-config is not a full replacement for the pkg-config tool, e.g.
// it does not implement conflict and dependency resolving.
//
// The cgo tool uses pkg-config for obtaining CFLAGS and LDFLAGS of C libraries.
// Example:
//
//   // #cgo pkg-config: libpng
//   // #include <png.h>
//   import "C"
//
// The "#cgo pkg-config: png" directive makes a cgo tool query a pkg-config for
// CFLAGS during generation of a C wrapper (--cflags flag) and for LDFLAGS during
// linking (--libs flag). The original pkg-config looks up a libpng.pc file in
// a directory list specified by the PKG_CONFIG_PATH environment variable plus
// a few other default ones. The libpng.pc file can have the following content:
//
//   prefix=/usr
//   exec_prefix=${prefix}
//   libdir=${prefix}/lib/x86_64-linux-gnu
//   includedir=${prefix}/include/libpng12
//
//   Name: libpng
//   Description: Loads and saves PNG files
//   Version: 1.2.46
//   Libs: -L${libdir} -lpng12
//   Libs.private: -lz -lm
//   Cflags: -I${includedir}
//
// A .pc file is composed of two parts - variables and keywords. Keywords may
// reference pre-declared variables.
//
//   $ pkg-config --cflags libpng
//   -I/usr/include/libpng12
//   $ pkg-config --libs libpng
//   -L/usr/lib/x86_64-linux-gnu -lpng12
//   $ pkg-config --cflags --libs libpng
//   -I/usr/include/libpng12 -L/usr/lib/x86_64-linux-gnu -lpng12
//
// The cmd/pkg-config tool looks up for a PC file in two other places in addition
// do the original pkg-config: $GOPATH and github.com.
//
// The cmd/pkg-config tool and $GOPATH
//
// The cmd/pkg-config defines standard directory layout for C libraries:
//
//   - $GOPATH/lib contains both the .pc files and binaries
//   - $GOPATH/include contains package headers
//
// The $GOPATH tree for the example may look like the following:
//
//   $GOPATH
//   ├── include
//   │   └── libpng
//   │       └── libpng12
//   │           ├── pngconf.h
//   │           └── png.h
//   ├── lib
//   │   ├── linux_amd64
//   │   │   └── libpng
//   │   │       ├── libpng12.so
//   │   │       └── libpng.pc
//   │   └── windows_amd64
//   │       └── libpng
//   │           ├── libpng12.dll
//   │           ├── libpng12.dll.a
//   │           └── libpng.pc
//   └── src
//       └── github.com
//           └── joe
//               └── png-wrapper
//                   └── png-wrapper.go
//
// The cmd/pkg-config reads the libpng.pc file from $GOPATH/lib/libpng/$GOOS_$GOARCH/libpng.pc.
// The .pc file written for cmd/pkg-config can use $GOPATH, $GOOS and $GOARCH
// builtin variables, which are expanded by the cmd/pkg-config during runtime.
// The rewritten .pc file for libpng may look like the following:
//
//   libdir=${GOPATH}/lib/${GOOS}_${GOARCH}/libpng
//   includedir=${GOPATH}/include/libpng
//
//   Name: libpng
//   Description: Loads and saves PNG files
//   Version: 1.2.46
//   Libs: -L${libdir} -lpng12
//   Libs.private: -lz -lm
//   Cflags: -I${includedir}
//
// The cmd/pkg-config tool and github.com
//
// Although it's advised to always use an official or self-compiled libraries for
// a production use, cmd/pkg-config can download a zip archive from project's
// github.com releases for a pkg-config tag and unpack it into $GOPATH.
// For example in order to make the above github.com/joe/png-wrapper package
// pkg-config-gettable, it's enough to zip include/ and lib/ directories:
//
//   $ zip -9 -r libpng.zip include dir
//
// Create release, name a tag after pkg-config and attach libpng.zip do the
// file list. This would make the libpng.zip archive be accessible from the following
// link:
//
//   $ wget http://github.com/joe/png-wrapper/releases/download/pkg-config/libpng.zip
//
// Which is the default location the cmd/pkg-config searches for libraries. Then
// go-getting a joe/png-wrapper package altogether with C dependencies is as
// easy as:
//
//   $ go get github.com/joe/png-wrapper
//
// The cmd/pkg-config tool looks up a .pc file for a $LIBRARY in the following order:
//
//   - $GOPATH/lib/$GOOS_$GOARCH/$LIBRARY/$LIBRARY.pc
//   - http://github.com/$USER/$PROJECT/releases/download/pkg-config/$LIBRARY.zip
//   - $PKG_CONFIG_PATH and eventual pkg-config's default search locations (platform-specific)
package main

import (
	"fmt"
	"os"

	"github.com/rjeczalik/pkgconfig"
)

const USAGE = `NAME:
		pkg-config - Go-centric pkg-config replacement

USAGE:
		pkg-config --libs PKG
		pkg-config --cflags PKG
		pkg-config --cflags --libs PKG1 PKG2
		pkg-config get github.com/USER/PROJ PKG`

func die(v ...interface{}) {
	for _, v := range v {
		fmt.Fprintln(os.Stderr, v)
	}
	os.Exit(1)
}

func ishelp(s string) bool {
	return s == "-h" || s == "-help" || s == "help" || s == "--help" || s == "/?"
}

func main() {
	if len(os.Args) == 1 || (len(os.Args) == 2 && ishelp(os.Args[1])) {
		fmt.Println(USAGE)
	} else {
		switch os.Args[1] {
		case "get":
			if len(os.Args) != 4 {
				die(USAGE)
			}
			pc, err := pkgconfig.LookupGithubProj(os.Args[3], os.Args[2])
			if err != nil {
				die(err)
			}
			_ = pc
		default:
			pkg := pkgconfig.NewPkgArgs(os.Args[1:])
			if err := pkg.Resolve(); err == nil {
				pkg.WriteTo(os.Stdout)
			} else {
				die(err)
			}
		}
	}
}
