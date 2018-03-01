# HP TRIM VMBX Email Format

This is a library and command line tool for working with HP TRIM VMBX email files. You can use this tool for extracting attachments and converting VMBX files to EML.

## Install

From source:

	go get github.com/srnsw/vmbx/cmd

Or get a precompiled Windows 64-bit binary from the [releases page](/releases).

To use the mail conversion function you also need to install a siegfried signature file on your computer (normally in a "siegfried" folder within your home directory). See [siegfried](https://github.com/richardlehane/siegfried) for more information.

## Usage

To extract attachments from a VMBX file, or set of files:

	vmbx -dump FILE or DIR

To convert a VMBX file, or set of files, to EML format, do:

	vmbx -mail FILE or DIR

If you haven't installed the siegfried signature file in a "siegfried" folder within your home directory, you can specify an alternate location with the "-sig" flag:

	vmbx -sig "c:\default.sig" -mail FILE or DIR


