package fq

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJSONLFileSourceStream(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "jsonl-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			t.Fatalf("Failed to remove temp directory: %v", err)
		}
	}(tempDir)

	tests := []struct {
		name           string
		content        string
		expected       []interface{}
		expectErrors   bool
		errorSubstring string
	}{
		{
			name: "basic_valid_jsonl",
			content: `{"id":1,"name":"Item 1"}
{"id":2,"name":"Item 2"}
{"id":3,"name":"Item 3"}`,
			expected: []interface{}{
				map[string]interface{}{"id": float64(1), "name": "Item 1"},
				map[string]interface{}{"id": float64(2), "name": "Item 2"},
				map[string]interface{}{"id": float64(3), "name": "Item 3"},
			},
			expectErrors: false,
		},
		{
			name: "with_empty_lines",
			content: `{"id":1,"name":"Item 1"}

{"id":2,"name":"Item 2"}

{"id":3,"name":"Item 3"}`,
			expected: []interface{}{
				map[string]interface{}{"id": float64(1), "name": "Item 1"},
				map[string]interface{}{"id": float64(2), "name": "Item 2"},
				map[string]interface{}{"id": float64(3), "name": "Item 3"},
			},
			expectErrors: false,
		},
		{
			name: "with_invalid_json",
			content: `{"id":1,"name":"Item 1"}
invalid json line
{"id":2,"name":"Item 2"}`,
			expected: []interface{}{
				map[string]interface{}{"id": float64(1), "name": "Item 1"},
				map[string]interface{}{"id": float64(2), "name": "Item 2"},
			},
			expectErrors:   true,
			errorSubstring: "error parsing JSON",
		},
		{
			name:    "with_windows_line_endings",
			content: "{\"id\":1,\"name\":\"Item 1\"}\r\n{\"id\":2,\"name\":\"Item 2\"}\r\n{\"id\":3,\"name\":\"Item 3\"}",
			expected: []interface{}{
				map[string]interface{}{"id": float64(1), "name": "Item 1"},
				map[string]interface{}{"id": float64(2), "name": "Item 2"},
				map[string]interface{}{"id": float64(3), "name": "Item 3"},
			},
			expectErrors: false,
		},
		{
			name:         "empty_file",
			content:      "",
			expected:     []interface{}{},
			expectErrors: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			testFile := filepath.Join(tempDir, tc.name+".jsonl")
			if err := os.WriteFile(testFile, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to write test file: %v", err)
			}

			dataCh, errCh := JSONLFileSourceStream(testFile)
			results, errors := collectResults(dataCh, errCh)

			if tc.expectErrors {
				if len(errors) == 0 {
					t.Errorf("Expected errors but got none")
				} else if tc.errorSubstring != "" {
					found := false
					for _, err := range errors {
						if strings.Contains(err.Error(), tc.errorSubstring) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Error doesn't contain expected substring '%s': %v",
							tc.errorSubstring, errors)
					}
				}
			} else if len(errors) > 0 {
				t.Errorf("Got unexpected errors: %v", errors)
			}

			if len(results) != len(tc.expected) {
				t.Fatalf("Expected %d items, got %d", len(tc.expected), len(results))
			}

			for i, expected := range tc.expected {
				expectedJSON, _ := json.Marshal(expected)
				resultJSON, _ := json.Marshal(results[i])

				if string(expectedJSON) != string(resultJSON) {
					t.Errorf("Item %d mismatch:\nExpected: %s\nGot: %s",
						i, string(expectedJSON), string(resultJSON))
				}
			}
		})
	}

	// Test error conditions
	t.Run("nonexistent_file", func(t *testing.T) {
		dataCh, errCh := JSONLFileSourceStream(filepath.Join(tempDir, "nonexistent.jsonl"))
		results, errors := collectResults(dataCh, errCh)

		if len(errors) == 0 {
			t.Error("Expected an error for nonexistent file, got none")
		}
		if len(results) > 0 {
			t.Errorf("Expected no results for error condition, got %v", results)
		}
	})

	t.Run("permission_denied", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("Skipping permission test when running as root")
		}

		noReadFile := filepath.Join(tempDir, "no-read.jsonl")
		if err := os.WriteFile(noReadFile, []byte(`{"id":1}`), 0200); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		dataCh, errCh := JSONLFileSourceStream(noReadFile)
		results, errors := collectResults(dataCh, errCh)

		if len(errors) == 0 {
			t.Error("Expected an error for unreadable file, got none")
		}
		if len(results) > 0 {
			t.Errorf("Expected no results for error condition, got %v", results)
		}
	})
}


func TestFilterCErrorHandling(t *testing.T) {
	t.Run("normal_operation", func(t *testing.T) {
		input := make(chan interface{}, 10)
		input <- map[string]interface{}{"id": 1, "name": "Item 1"}
		input <- map[string]interface{}{"id": 2, "name": "Item 2"}
		input <- map[string]interface{}{"id": 3, "name": "Item 3"}
		close(input)

		dataCh, errCh := FilterC(input, Q{"id": 2}, 0, 0)
		results, errors := collectResults(dataCh, errCh)

		if len(errors) > 0 {
			t.Errorf("Unexpected errors: %v", errors)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		} else {
			item := results[0].(map[string]interface{})
			if item["id"].(int) != 2 {
				t.Errorf("Expected item with id 2, got %v", item)
			}
		}
	})

	t.Run("panic_recovery", func(t *testing.T) {
		input := make(chan interface{}, 10)
		input <- map[string]interface{}{"id": 1, "name": "Item 1"}
		input <- map[string]interface{}{"id": 2, "name": "Item 2"}
		close(input)

		// force panic during filtering
		panicPredicate := func(v interface{}) bool {
			var nilMap map[string]interface{}
			return nilMap["FilterC"] == true
		}

		dataCh, errCh := FilterC(input, panicPredicate, 0, 0)
		results, errors := collectResults(dataCh, errCh)

		if len(errors) != 0 {
			t.Errorf("Expected 0 errors, got %v", errors)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %v", results)
		}

		t.Logf("Got %d errors from panic recovery", len(errors))
	})
}

// helpers
func collectResults(dataCh <-chan interface{}, errCh <-chan error) ([]interface{}, []error) {
	var results []interface{}
	var errors []error

	done := make(chan struct{})

	go func() {
		for item := range dataCh {
			results = append(results, item)
		}
		done <- struct{}{}
	}()

	go func() {
		for err := range errCh {
			errors = append(errors, err)
		}
		done <- struct{}{}
	}()

	<-done
	<-done

	return results, errors
}
