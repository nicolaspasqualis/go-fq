package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Build the CLI binary for testing
	if err := exec.Command("go", "build", "-o", "testfq", ".").Run(); err != nil {
		panic("Failed to build CLI for testing: " + err.Error())
	}
	
	code := m.Run()
	
	// Clean up
	os.Remove("testfq")
	os.Exit(code)
}

func createTestData(t *testing.T) string {
	content := `{"name": "laptop", "price": 999.99, "category": "electronics", "tags": ["portable", "work"]}
{"name": "book", "price": 29.99, "category": "books", "tags": ["education", "paperback"]}
{"name": "smartphone", "price": 699.99, "category": "electronics", "tags": ["mobile", "communication"]}
{"name": "desk", "price": 299.99, "category": "furniture", "tags": ["office", "wooden"]}
{"name": "headphones", "price": 199.99, "category": "electronics", "tags": ["audio", "wireless"]}
{"name": "chair", "price": 150.00, "category": "furniture", "tags": ["office", "comfortable"]}
`
	
	file, err := os.CreateTemp("", "test-data-*.jsonl")
	if err != nil {
		t.Fatal("Failed to create test file:", err)
	}
	
	if _, err := file.WriteString(content); err != nil {
		t.Fatal("Failed to write test data:", err)
	}
	file.Close()
	
	return file.Name()
}

func runCLI(args ...string) (stdout, stderr string, exitCode int) {
	cmd := exec.Command("./testfq", args...)
	
	stdoutBytes, err := cmd.Output()
	stdout = string(stdoutBytes)
	
	if exitErr, ok := err.(*exec.ExitError); ok {
		stderr = string(exitErr.Stderr)
		exitCode = exitErr.ExitCode()
	} else if err != nil {
		stderr = err.Error()
		exitCode = 1
	}
	
	return stdout, stderr, exitCode
}

func TestBasicFiltering(t *testing.T) {
	testFile := createTestData(t)
	defer os.Remove(testFile)
	
	tests := []struct {
		name     string
		args     []string
		wantExit int
		contains []string
		notContains []string
	}{
		{
			name:     "filter by price greater than 500",
			args:     []string{testFile, "price:gt:500"},
			wantExit: 0,
			contains: []string{"laptop", "smartphone"},
			notContains: []string{"book", "chair"},
		},
		{
			name:     "filter by category electronics",
			args:     []string{testFile, "category:eq:electronics"},
			wantExit: 0,
			contains: []string{"laptop", "smartphone", "headphones"},
			notContains: []string{"book", "desk", "chair"},
		},
		{
			name:     "filter by tags containing work",
			args:     []string{testFile, "tags:hasitem:work"},
			wantExit: 0,
			contains: []string{"laptop"},
			notContains: []string{"book", "smartphone"},
		},
		{
			name:     "filter with IN operator",
			args:     []string{testFile, "category:in:electronics,furniture"},
			wantExit: 0,
			contains: []string{"laptop", "desk", "chair"},
			notContains: []string{"book"},
		},
		{
			name:     "multiple filters",
			args:     []string{testFile, "category:eq:electronics", "price:lt:500"},
			wantExit: 0,
			contains: []string{"headphones"},
			notContains: []string{"laptop", "smartphone", "book"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runCLI(tt.args...)
			
			if exitCode != tt.wantExit {
				t.Errorf("Expected exit code %d, got %d. Stderr: %s", tt.wantExit, exitCode, stderr)
			}
			
			for _, want := range tt.contains {
				if !strings.Contains(stdout, want) {
					t.Errorf("Expected output to contain %q, but it didn't. Output: %s", want, stdout)
				}
			}
			
			for _, notWant := range tt.notContains {
				if strings.Contains(stdout, notWant) {
					t.Errorf("Expected output to not contain %q, but it did. Output: %s", notWant, stdout)
				}
			}
		})
	}
}

func TestOutputFormats(t *testing.T) {
	testFile := createTestData(t)
	defer os.Remove(testFile)
	
	tests := []struct {
		name     string
		args     []string
		wantExit int
		check    func(t *testing.T, stdout string)
	}{
		{
			name:     "jsonl_format",
			args:     []string{testFile, "category:eq:electronics"},
			wantExit: 0,
			check: func(t *testing.T, stdout string) {
				lines := strings.Split(strings.TrimSpace(stdout), "\n")
				for _, line := range lines {
					if line != "" && (!strings.HasPrefix(line, "{") || !strings.HasSuffix(line, "}")) {
						t.Errorf("JSONL line should be valid JSON: %s", line)
					}
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runCLI(tt.args...)
			
			if exitCode != tt.wantExit {
				t.Errorf("Expected exit code %d, got %d. Stderr: %s", tt.wantExit, exitCode, stderr)
			}
			
			tt.check(t, stdout)
		})
	}
}

func TestSkipAndLimit(t *testing.T) {
	testFile := createTestData(t)
	defer os.Remove(testFile)
	
	tests := []struct {
		name        string
		args        []string
		wantExit    int
		expectLines int  // number of data lines (excluding header)
	}{
		{
			name:        "limit to 2 results",
			args:        []string{"-limit", "2", testFile, "price:gt:0"},
			wantExit:    0,
			expectLines: 2,
		},
		{
			name:        "skip first result",
			args:        []string{"-skip", "1", "-limit", "2", testFile, "price:gt:0"},
			wantExit:    0,
			expectLines: 2,
		},
		{
			name:        "skip more than available",
			args:        []string{"-skip", "100", testFile, "price:gt:0"},
			wantExit:    0,
			expectLines: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runCLI(tt.args...)
			
			if exitCode != tt.wantExit {
				t.Errorf("Expected exit code %d, got %d. Stderr: %s", tt.wantExit, exitCode, stderr)
			}
			
			lines := strings.Split(strings.TrimSpace(stdout), "\n")
			dataLines := 0
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					dataLines++
				}
			}
			
			if dataLines != tt.expectLines {
				t.Errorf("Expected %d data lines, got %d. Output: %s", tt.expectLines, dataLines, stdout)
			}
		})
	}
}

func TestErrorHandling(t *testing.T) {
	testFile := createTestData(t)
	defer os.Remove(testFile)
	
	tests := []struct {
		name     string
		args     []string
		wantExit int
		stderrContains string
	}{
		{
			name:           "missing file",
			args:           []string{"nonexistent.jsonl", "price:gt:100"},
			wantExit:       1,
			stderrContains: "no such file",
		},
		{
			name:           "invalid operator",
			args:           []string{testFile, "price:invalid:100"},
			wantExit:       1,
			stderrContains: "unknown operator",
		},
		{
			name:           "invalid filter format",
			args:           []string{testFile, "price-gt-100"},
			wantExit:       1,
			stderrContains: "invalid filter format",
		},
		{
			name:           "missing data file",
			args:           []string{},
			wantExit:       1,
			stderrContains: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runCLI(tt.args...)
			
			if exitCode != tt.wantExit {
				t.Errorf("Expected exit code %d, got %d. Stdout: %s, Stderr: %s", tt.wantExit, exitCode, stdout, stderr)
			}
			
			if tt.stderrContains != "" && !strings.Contains(stderr, tt.stderrContains) {
				t.Errorf("Expected stderr to contain %q, got: %s", tt.stderrContains, stderr)
			}
		})
	}
}

func TestComplexFilters(t *testing.T) {
	testFile := createTestData(t)
	defer os.Remove(testFile)
	
	tests := []struct {
		name     string
		args     []string
		wantExit int
		contains []string
	}{
		{
			name:     "in operator for field value",
			args:     []string{testFile, "category:in:books,furniture"},
			wantExit: 0,
			contains: []string{"book", "desk", "chair"},
		},
		{
			name:     "variadic function with multiple values",
			args:     []string{testFile, "tags:containsany:work,audio,office"},
			wantExit: 0,
			contains: []string{"laptop", "headphones", "desk", "chair"},
		},
		{
			name:     "numeric comparisons",
			args:     []string{testFile, "price:gte:200"},
			wantExit: 0,
			contains: []string{"laptop", "smartphone", "desk"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runCLI(tt.args...)
			
			if exitCode != tt.wantExit {
				t.Errorf("Expected exit code %d, got %d. Stderr: %s", tt.wantExit, exitCode, stderr)
			}
			
			for _, want := range tt.contains {
				if !strings.Contains(stdout, want) {
					t.Errorf("Expected output to contain %q, but it didn't. Output: %s", want, stdout)
				}
			}
		})
	}
}

func TestHelpAndUsage(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantExit int
		contains []string
	}{
		{
			name:     "help flag",
			args:     []string{"-help"},
			wantExit: 0,
			contains: []string{"Usage:", "fq", "Options:", "Examples:"},
		},
		{
			name:     "no arguments shows usage",
			args:     []string{},
			wantExit: 1,
			contains: []string{"Usage:", "fq"},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runCLI(tt.args...)
			output := stdout + stderr
			
			if exitCode != tt.wantExit {
				t.Errorf("Expected exit code %d, got %d", tt.wantExit, exitCode)
			}
			
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("Expected output to contain %q, but it didn't. Output: %s", want, output)
				}
			}
		})
	}
}