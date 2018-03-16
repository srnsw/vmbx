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

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/srnsw/vmbx"
)

var (
	dflag = flag.Bool("dump", false, "dump attachments from VMBX file or VMBX files within directory")
	mflag = flag.Bool("mail", false, "convert VMBX file or VMBX files within directory to multipart MIME email")
	sfLoc = flag.String("sig", "", "location of siegfried signature file")
)

func main() {
	flag.Parse()
	target := flag.Arg(0)
	if target == "" {
		log.Fatal("need a VMBX file or directory target")
	}
	if *mflag {
		if *sfLoc == "" {
			usr, err := user.Current()
			if err == nil {
				*sfLoc = filepath.Join(usr.HomeDir, "siegfried", "default.sig")
			}
		}
		if _, err := os.Stat(*sfLoc); err != nil {
			log.Fatal("can't find location of siegfried signature file")
		}
	}

	err := filepath.Walk(target, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.ToUpper(filepath.Ext(path)) == ".VMBX" {
			if *dflag {
				if err = dump(path); err != nil {
					return fmt.Errorf("Error dumping %s, got: %v", path, err)
				}
			}
			if *mflag {
				if err = mail(path); err != nil {
					return fmt.Errorf("Error converting %s to EML, got: %v", path, err)
				}
			}
		}
		return nil
	})
	log.Print(err)
}

func mail(p string) error {
	f, err := os.Open(p)
	defer f.Close()
	if err != nil {
		return err
	}
	v, err := vmbx.New(f)
	if err != nil {
		return nil // not vmbx
	}
	r := strings.NewReplacer(".vmbx", ".eml", ".VMBX", ".EML")
	fname := r.Replace(filepath.Base(p))
	target, err := os.Create(fname)
	if err != nil {
		return err
	}
	err = v.Mail(target, *sfLoc)
	if err != nil {
		return err
	}
	return target.Close()
}

func dump(p string) error {
	f, err := os.Open(p)
	defer f.Close()
	if err != nil {
		return err
	}
	v, err := vmbx.New(f)
	if err != nil {
		return nil // not vmbx
	}
	as := v.Attachments()
	if len(as) == 0 {
		return nil // no attach
	}
	dirname := filepath.Join(filepath.Dir(p), strings.Replace(filepath.Base(p), ".", "_", -1)+"_attach")
	err = os.MkdirAll(dirname, os.ModeDir)
	if err != nil {
		return err
	}
	for _, a := range as {
		fname, err := os.Create(filepath.Join(dirname, a.Name))
		if err != nil {
			return err
		}
		rdr, err := a.Reader(true)
		if err != nil {
			return err
		}
		_, err = io.Copy(fname, rdr)
		if err != nil {
			return err
		}
		fname.Close()
	}
	return nil
}
