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
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"strings"
	"time"

	"github.com/richardlehane/siegfried"
	"github.com/richardlehane/siegfried/pkg/pronom"
)

var (
	sf *siegfried.Siegfried

	dateFmt1 = "02/01/2006 15:04:05"
	dateFmt2 = "02-Jan-2006 15:04:05"
)

func (v *VMBX) Mail(w io.Writer, sfLoc string) error {
	var err error
	if sf == nil {
		sf, err = siegfried.Load(sfLoc)
		if err != nil {
			return err
		}
	}
	uniqs := map[string]bool{
		"TRIM-Embedded": true,
	}
	for _, k := range v.Keys {
		if uniqs[k] {
			continue
		}
		uniqs[k] = true
		fmt.Fprintf(w, "%s: %s\r\n", k, strings.Join(v.Headers[k], ","))
	}
	var sentDate, sentTime string
	if len(v.Headers["Sent-Date"]) > 0 {
		sentDate = v.Headers["Sent-Date"][0]
	}
	if len(v.Headers["Sent-Time"]) > 0 {
		sentTime = v.Headers["Sent-Time"][0]
	}
	t, err := time.Parse(dateFmt1, sentDate+" "+sentTime)
	if err != nil {
		t, err = time.Parse(dateFmt2, sentDate+" "+sentTime)
	}
	if err == nil {
		fmt.Fprintf(w, "Date: %s\r\n", t.Format(time.RFC822))
	}
	m := multipart.NewWriter(w)
	fmt.Fprintf(w, "MIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", m.Boundary())

	bodyHdr := make(textproto.MIMEHeader)
	bodyHdr.Set("Content-Type", "text/plain")
	bw, err := m.CreatePart(bodyHdr)
	if err != nil {
		return err
	}
	body, err := v.Body()
	if err != nil {
		return err
	}
	_, err = io.Copy(bw, body)
	if err != nil {
		return err
	}
	attachHdr := make(textproto.MIMEHeader)
	attachHdr.Add("Content-Transfer-Encoding", "base64")
	for _, a := range v.Attachments() {
		attachHdr.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", a.Name))
		irdr, err := a.Reader(true)
		ids, err := sf.Identify(irdr, a.Name, "")
		if err != nil {
			return fmt.Errorf("Error identifying %s; got %v", a.Name, err)
		}
		var mime string
		for _, id := range ids {
			pid, ok := id.(pronom.Identification)
			if ok {
				mime = pid.MIME
				break
			}
		}
		if mime == "" {
			mime = "application/octet-stream"
		}
		attachHdr.Set("Content-Type", mime)
		aw, err := m.CreatePart(attachHdr)
		if err != nil {
			return err
		}
		rdr, err := a.Reader(false)
		if err != nil {
			return err
		}
		_, err = io.Copy(aw, rdr)
		if err != nil {
			return err
		}
	}
	return m.Close()
}
