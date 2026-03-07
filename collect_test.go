package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/sirkon/errors"
)

func TestCollect(t *testing.T) {
	ldata, err := os.ReadFile("testdata/linuxbench.txt")
	if err != nil {
		t.Fatal(errors.Wrap(err, "read data for linux bench"))
	}

	lcoll, err := Collect(bytes.NewReader(ldata))
	if err != nil {
		t.Fatal(errors.Wrap(err, "collect data for linux bench"))
	}

	if len(lcoll.Results) == 0 {
		t.Fatal("no bench results for linux collected")
	}

	mdata, err := os.ReadFile("testdata/macbench.txt")
	if err != nil {
		t.Fatal(errors.Wrap(err, "read data for macos bench"))
	}

	mcoll, err := Collect(bytes.NewReader(mdata))
	if err != nil {
		t.Fatal(errors.Wrap(err, "collect data for macos bench"))
	}

	if len(mcoll.Results) == 0 {
		t.Fatal("no bench results for macos collected")
	}

	if len(lcoll.Results) != len(mcoll.Results) {
		t.Fatalf("collected different bench sets for linux and macos: %d != %d", len(lcoll.Results), len(mcoll.Results))
	}

	for i := range lcoll.Results {
		lname := lcoll.Results[i].Name
		mname := mcoll.Results[i].Name
		if strings.HasPrefix(mname, lname) || strings.HasPrefix(lname, mname) {
			continue
		}

		t.Errorf("linux and macos mismatch: %q != %q", lname, mname)
	}
}
