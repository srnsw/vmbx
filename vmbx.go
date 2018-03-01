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
	"bufio"
	"encoding/base64"
	"errors"
	"io"
	"strconv"
	"strings"
)

type VMBX struct {
	Headers map[string][]string
	Keys    []string

	src io.ReadSeeker
}

type Attachment struct {
	Name          string
	Extension     string
	Size          int64
	EncodedOffset int64
	EncodedSize   int64

	vmbx *VMBX
}

func (a *Attachment) Reader(decode bool) (io.Reader, error) {
	_, err := a.vmbx.src.Seek(a.EncodedOffset, io.SeekStart)
	if err != nil {
		return nil, err
	}
	lr := io.LimitReader(a.vmbx.src, a.EncodedSize)
	if !decode {
		return lr, nil
	}
	return base64.NewDecoder(base64.StdEncoding, lr), nil
}

func mustInt(s string, b int) int64 {
	i, _ := strconv.ParseInt(s, b, 64)
	return i
}

func (v *VMBX) Attachments() []*Attachment {
	te, ok := v.Headers["TRIM-Embedded"]
	if !ok {
		return nil
	}
	ret := make([]*Attachment, len(te))
	for i, a := range te {
		m := splitTE(a)
		ret[i] = &Attachment{
			Name:          m["Name"],
			Extension:     m["Extension"],
			Size:          mustInt(m["Size"], 10),
			EncodedOffset: mustInt(m["EncodedOffset"], 16),
			EncodedSize:   mustInt(m["EncodedSize"], 10),
			vmbx:          v,
		}
	}
	return ret
}

func splitTE(s string) map[string]string {
	ret := make(map[string]string)
	ss := strings.Split(s, "\",")
	for _, v := range ss {
		kv := strings.SplitN(v, "=\"", 2)
		if len(kv) != 2 {
			return nil
		}
		ret[kv[0]] = strings.TrimSuffix(kv[1], "\"")
	}
	return ret
}

func (v *VMBX) Body() (io.Reader, error) {
	var headerLen int64 = 2 // final carriage + line return
	for k, vals := range v.Headers {
		for _, val := range vals {
			headerLen += int64(len(k) + len(val) + 3) // add the colon and the carriage + line return
		}
	}
	_, err := v.src.Seek(headerLen, io.SeekStart)
	if err != nil {
		return nil, err
	}
	as := v.Attachments()
	if len(as) == 0 {
		return v.src, nil
	}
	var firstAttachOff int64
	var firstAttachName string
	for _, a := range as {
		if firstAttachOff > 0 && firstAttachOff < a.EncodedOffset {
			continue
		}
		firstAttachOff = a.EncodedOffset
		firstAttachName = a.Name
	}
	sz := firstAttachOff - 19 - int64(len(firstAttachName)) - headerLen
	if sz < 0 {
		sz = 0
	}
	return io.LimitReader(v.src, sz), nil
}

func New(r io.ReadSeeker) (*VMBX, error) {
	scanner := bufio.NewScanner(r)
	v := &VMBX{
		Headers: make(map[string][]string),
		Keys:    make([]string, 0, 10),
		src:     r,
	}
	for scanner.Scan() {
		if scanner.Text() == "" {
			return v, nil
		}
		kv := strings.SplitN(scanner.Text(), ":", 2)
		if len(kv) != 2 {
			break
		}
		v.Headers[kv[0]] = append(v.Headers[kv[0]], kv[1])
		v.Keys = append(v.Keys, kv[0])
	}
	return nil, errors.New("vmbx: Unable to parse as VMBX file")
}
