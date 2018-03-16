// Copyright 2018 State of New South Wales through the State Records Authority of NSW. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vmbx

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

var sfLoc string

func init() {
	usr, err := user.Current()
	if err == nil {
		sfLoc = filepath.Join(usr.HomeDir, "siegfried", "default.sig")
	}
}

func TestNew(t *testing.T) {
	if _, err := os.Stat("test.vmbx"); err != nil {
		return
	}
	f, err := os.Open("test.vmbx")
	defer f.Close()
	if err != nil {
		t.Fatalf("failed to open test file got %v", err)
	}
	v, err := New(f)
	if err != nil {
		t.Fatalf("failed to parse vmbx file got %v", err)
	}
	for _, a := range v.Attachments() {
		rdr, err := a.Reader(true)
		if err != nil {
			t.Fatalf("failed to get reader got %v", err)
		}
		byts, err := ioutil.ReadAll(rdr)
		if err != nil {
			t.Fatalf("bad read got %v", err)
		}
		if int64(len(byts)) != a.Size {
			t.Fatalf("expected bytes to equal file size got %d bytes for %d size", len(byts), a.Size)
		}
	}
}

func TestMail(t *testing.T) {
	if _, err := os.Stat("test.vmbx"); err != nil {
		return
	}
	f, err := os.Open("test.vmbx")
	defer f.Close()
	if err != nil {
		t.Fatalf("failed to open test file got %v", err)
	}
	v, err := New(f)
	if err != nil {
		t.Fatalf("failed to parse vmbx file got %v", err)
	}
	buf := &bytes.Buffer{}
	err = v.Mail(buf, sfLoc)
	if err != nil {
		t.Fatalf("failed to write mail message, got %v", err)
	}
}

func TestEmpty(t *testing.T) {
	if _, err := os.Stat("test_empty.vmbx"); err != nil {
		return
	}
	f, err := os.Open("test_empty.vmbx")
	defer f.Close()
	if err != nil {
		t.Fatalf("failed to open test file got %v", err)
	}
	v, err := New(f)
	if err != nil {
		t.Fatalf("failed to parse vmbx file got %v", err)
	}
	buf := &bytes.Buffer{}
	err = v.Mail(buf, sfLoc)
	if err != nil {
		t.Fatalf("failed to write mail message, got %v", err)
	}
}
