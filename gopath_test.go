package pkgconfig

import (
	"os"
	"testing"
)

func TestLookupGopath(t *testing.T) {
	if err := os.Setenv("GOPATH", "testdata"); err != nil {
		t.Fatalf("expected err=nil; was %q", err)
	}
	pc, err := LookupGopath("libgit2")
	if err != nil {
		t.Fatalf("expected err=nil; was %q", err)
	}
	if pc.Name != "libgit2" {
		t.Errorf(`expected pc.Name="libgit2"; was %q`, pc.Name)
	}
	if pc.Desc != "The git library, take 2" {
		t.Errorf(`expected pc.Desc="The git library, take"; was %q`, pc.Desc)
	}
	if pc.Version != "0.20.0" {
		t.Errorf(`expected pc.Version="0.20.0"; was %q`, pc.Version)
	}
	if len(pc.Libs) == 0 {
		t.Errorf("expected len(pc.Libs)!=0")
	}
	if len(pc.Cflags) == 0 {
		t.Errorf("expected len(pc.Cflags)!=0")
	}
}
