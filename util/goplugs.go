package util

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"rpg/files"
)

func MustOpen(filename string) io.ReaderAt {
	f, err := files.Open(filename)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to open file '%s': %w", filename, err))
	}
	readerAt, ok := f.(io.ReaderAt)
	if !ok {
		return nil
	}
	return readerAt
}

func MustOpenReader(filename string) io.ReadCloser {
	f, err := files.Open(filename)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to open file '%s': %w", filename, err))
	}
	return f
}

func ReadAllLines(filename string) []string {
	file := MustOpenReader(filename)
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var fileLines []string
	for scanner.Scan() {
		line := scanner.Text()
		fileLines = append(fileLines, line)
	}
	return fileLines
}