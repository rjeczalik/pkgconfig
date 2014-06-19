package pkgconfig

import (
	"bytes"
	"io"
	"strings"
)

var lookups = map[string]func(string) (*PC, error){
	"$GOPATH":          LookupGopath,
	"github.com":       LookupGithub,
	"$PKG_CONFIG_PATH": LookupPC,
}

type namedErrors map[string]error

func (ne namedErrors) Error() string {
	s := make([]string, 0, len(ne))
	for name, err := range ne {
		s = append(s, "error "+name+": "+err.Error())
	}
	return strings.Join(s, "\n")
}

// DefaultLookup TODO(rjeczalik): document
func DefaultLookup(pkg string) (pc *PC, err error) {
	ne := make(map[string]error, len(lookups))
	for name, lookup := range lookups {
		if pc, err = lookup(pkg); err == nil {
			return
		}
		ne[name] = err
	}
	return nil, namedErrors(ne)
}

// Pkg TODO(rjeczalik): document
type Pkg struct {
	Packages []string
	Libs     bool
	Cflags   bool
	Lookup   func(string) (*PC, error)
	pc       []*PC
}

// NewPkgArgs TODO(rjeczalik): document
func NewPkgArgs(args []string) *Pkg {
	pkg := &Pkg{}
	for _, arg := range args {
		switch {
		case arg == "--libs":
			pkg.Libs = true
		case arg == "--cflags":
			pkg.Cflags = true
		case strings.HasPrefix(arg, "-"):
		default:
			pkg.Packages = append(pkg.Packages, arg)
		}
	}
	return pkg
}

// Resolve TODO(rjeczalik): document
func (pkg *Pkg) Resolve() error {
	if len(pkg.Packages) == 0 {
		return ErrEmptyPC
	}
	lu := pkg.Lookup
	if lu == nil {
		lu = DefaultLookup
	}
	var (
		pc   = make([]*PC, 0, len(pkg.Packages))
		dups = make(map[string]struct{})
	)
	for _, p := range pkg.Packages {
		if _, ok := dups[p]; !ok {
			pkg, err := lu(p)
			if err != nil {
				return err
			}
			pc = append(pc, pkg)
			dups[p] = struct{}{}
		}
	}
	pkg.pc = pc
	return nil
}

// WriteTo TODO(rjeczalik): document
func (pkg Pkg) WriteTo(w io.Writer) (int64, error) {
	var (
		dups = make(map[string]struct{})
		buf  bytes.Buffer
	)
	if pkg.Cflags {
		for _, pc := range pkg.pc {
			for _, cflag := range pc.Cflags {
				if _, ok := dups[cflag]; !ok {
					buf.WriteString(cflag)
					buf.WriteByte(' ')
					dups[cflag] = struct{}{}
				}
			}
		}
	}
	if pkg.Libs {
		for _, pc := range pkg.pc {
			for _, lib := range pc.Libs {
				if _, ok := dups[lib]; !ok {
					buf.WriteString(lib)
					buf.WriteByte(' ')
					dups[lib] = struct{}{}
				}
			}
		}
	}
	if buf.Len() == 0 {
		return 0, ErrEmptyPC
	}
	p := buf.Bytes()
	p[len(p)-1] = '\n'
	return io.Copy(w, bytes.NewBuffer(p))
}
