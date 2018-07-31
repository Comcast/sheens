package tools

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

// Inline replaces '%inline("NAME")' with f(NAME).
func Inline(bs []byte, f func(string) ([]byte, error)) ([]byte, error) {
	p, err := regexp.Compile(`(?s)(.*?)(%inline *\("([^"]*)"\))`)
	if err != nil {
		return nil, err
	}
	i := 0
	acc := make([]byte, 0, len(bs))
	for {
		part := p.FindSubmatch(bs[i:])
		if part == nil {
			acc = append(acc, bs[i:]...)
			break
		}
		i += len(part[0])
		acc = append(acc, part[1]...)
		replacement, err := f(string(part[3]))
		if err != nil {
			return nil, err
		}
		log.Printf("debug inlining %s: %s", part[3], replacement)
		acc = append(acc, replacement...)
	}

	return acc, nil
}

// ReadFileWithInlines is a replacement for ioutil.ReadFile that adds
// automation Inline()ing based on the directory obtained from the
// filename.
//
// '%inline("NAME")' is replaced with ReadFile(NAME).
func ReadFileWithInlines(filename string) ([]byte, error) {

	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(filename)
	f := func(name string) ([]byte, error) {
		return ioutil.ReadFile(dir + string(os.PathSeparator) + name)
	}

	return Inline(bs, f)
}

// ReadFileWithInlines is a replacement for ioutil.ReadAll that adds
// automation Inline()ing based on the given directory.
//
// '%inline("NAME")' is replaced with ReadFile(NAME).
func ReadAllWithInlines(in io.Reader, dir string) ([]byte, error) {

	bs, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}

	f := func(name string) ([]byte, error) {
		return ioutil.ReadFile(dir + string(os.PathSeparator) + name)
	}

	return Inline(bs, f)
}
