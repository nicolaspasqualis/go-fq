package fq

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// JSONLFileSourceStream creates a channel of objects parsed from a JSONL file and a channel for errors.
func JSONLFileSourceStream(filePath string) (<-chan interface{}, <-chan error) {
	output := make(chan interface{}, 100)
	errCh := make(chan error, 10)

	file, err := os.Open(filePath)
	if err != nil {
		go func() {
			errCh <- fmt.Errorf("failed to open file: %w", err)
			close(errCh)
			close(output)
		}()
		return output, errCh
	}

	go func() {
		defer file.Close()
		defer close(output)
		defer close(errCh)

		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			if strings.TrimSpace(line) == "" {
				continue
			}

			var obj interface{}
			if err := json.Unmarshal([]byte(line), &obj); err != nil {
				errCh <- fmt.Errorf("line %d: error parsing JSON: %w", lineNum, err)
				continue
			}

			output <- obj
		}

		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("error reading file: %w", err)
		}
	}()

	return output, errCh
}

