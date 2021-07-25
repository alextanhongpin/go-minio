package main

import (
	"fmt"
	"log"
	"mime"
	"path/filepath"
	"regexp"
	"strings"
)

var re *regexp.Regexp

func init() {
	var err error
	re, err = regexp.Compile(`^[a-z0-9][\w-]+[a-z0-9]$`)
	if err != nil {
		log.Fatalln(err)
	}
}

func ValidateFilename(value string) bool {
	return re.MatchString(value)
}

// Filename holds information for the filename.
type Filename struct {
	name        string
	ext         string
	contentType string
}

func NewFilename(file string) (Filename, error) {
	ext := filepath.Ext(file)
	name := file[:len(file)-len(ext)]
	if !ValidateFilename(name) {
		return Filename{}, fmt.Errorf("invalid filename %q: file name accepts only characters a-z, or 0-9, a '-' or '_'", name)
	}

	contentType := mime.TypeByExtension(ext)
	if !strings.HasPrefix(contentType, "image/") {
		return Filename{}, fmt.Errorf("%q is not a valid image extension", ext)
	}

	return Filename{
		name:        name,
		ext:         ext,
		contentType: contentType,
	}, nil
}

func (f Filename) Path() string {
	return f.name + f.ext
}

// Name returns the filename without extension.
func (f Filename) Name() string {
	return f.name
}

// Extension returns extension with the prefix dot '.',
// e.g. '.png'.
func (f Filename) Extension() string {
	return f.ext
}

// ContentType returns the file Content-Type, e.g.
// image/png, image/svg+xml.
func (f Filename) ContentType() string {
	return f.contentType
}
