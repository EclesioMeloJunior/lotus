package source

import (
	"fmt"
	"os"
	"path/filepath"
)

var allowedExtensions = map[string]bool{
	".lt": true,
}

type SourceFile struct {
	source   []byte
	cursorAt uint
	path     string
}

func FromFile(sourcePath string) (*SourceFile, error) {
	ext := filepath.Ext(sourcePath)
	if !allowedExtensions[ext] {
		return nil, fmt.Errorf("file extension %s is not allowed", ext)
	}

	contents, err := os.ReadFile(sourcePath)
	if err != nil {
		return nil, err
	}
	return &SourceFile{source: contents, path: sourcePath}, nil
}

func (s *SourceFile) Read(p []byte) (n int, err error) {
	n = copy(p, s.source[s.cursorAt:])
	s.cursorAt += uint(n)
	return n, nil
}
