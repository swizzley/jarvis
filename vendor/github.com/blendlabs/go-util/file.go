package util

import (
	"bufio"
	"io"
	"os"

	"github.com/blendlabs/go-exception"
)

type ReadChunkHandler func(line []byte)
type ReadLineHandler func(line string)

func ReadFileByLines(filePath string, handler ReadLineHandler) error {
	if f, err := os.Open(filePath); err == nil {
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			handler(line)
		}
	} else {
		return exception.Wrap(err)
	}
	return nil
}

func ReadFileByChunks(filePath string, chunkSize int, handler ReadChunkHandler) error {
	if f, err := os.Open(filePath); err == nil {
		defer f.Close()

		chunk := make([]byte, chunkSize)
		for {
			readBytes, err := f.Read(chunk)
			if err == io.EOF {
				break
			}
			readData := chunk[:readBytes]
			handler(readData)
		}
	} else {
		return exception.Wrap(err)
	}
	return nil
}
