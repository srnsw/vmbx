# HP TRIM VMBX Email tool

This is a golang library and command line tool for extracting attachments from, and converting to EML, HP TRIM VMBX email files.

## Install

	go get -o vmbx.exe github.com/srnsw/vmbx/cmd

To use the mail conversion function you also need to install a siegfried signature file on your computer (normally in a "siegfried" folder within your home directory). See github.com/richardlehane/siegfried for more information.

## Usage

To extract attachments from a VMBX file, or set of files:

	vmbx -dump FILE or DIR

To convert a VMBX file, or set of files, to multipart-MIME encoded EML format, do:

	vmbx -mail FILE or DIR

If you haven't installed the siegfried signature file in a "siegfried" folder within your home directory, you can specify an alternate location with the "-sig" flag:

	vmbx -sig "c:\default.sig" -mail FILE or DIR


