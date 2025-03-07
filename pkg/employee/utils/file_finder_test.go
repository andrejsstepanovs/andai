package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/andrejsstepanovs/andai/pkg/employee/utils"
	"github.com/stretchr/testify/assert"
)

func TestFindFilesInText(t *testing.T) {
	// Setup test files
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"simple.txt",
		"with space.txt",
		"nested/file.go",
		"nested/deeper/config.yaml",
	}

	for _, file := range testFiles {
		path := filepath.Join(tempDir, file)
		dir := filepath.Dir(path)

		// Create directory if needed
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Create empty file
		f, err := os.Create(path)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		f.Close()
	}

	// Get absolute paths for verification
	simpleAbsPath, _ := filepath.Abs(filepath.Join(tempDir, "simple.txt"))
	spaceAbsPath, _ := filepath.Abs(filepath.Join(tempDir, "with space.txt"))
	nestedAbsPath, _ := filepath.Abs(filepath.Join(tempDir, "nested/file.go"))
	deeperAbsPath, _ := filepath.Abs(filepath.Join(tempDir, "nested/deeper/config.yaml"))

	// Fix the struct usage
	finder := utils.NewFileFinder([]string{tempDir, ""})

	tests := []struct {
		name     string
		content  string
		expected []utils.FileInfo
	}{
		{
			name:     "Empty text",
			content:  "",
			expected: []utils.FileInfo{},
		},
		{
			name:     "No valid paths",
			content:  "This text contains no valid file paths.",
			expected: []utils.FileInfo{},
		},
		{
			name:    "Single unquoted path",
			content: "Check this file: " + filepath.Join(tempDir, "simple.txt"),
			expected: []utils.FileInfo{
				{
					Original: filepath.Join(tempDir, "simple.txt"),
					Resolved: simpleAbsPath,
				},
			},
		},
		{
			name:    "Single unquoted path in search directories",
			content: "<error>\n# github.bus.zalan.do/aaa/bbb/pkg/repository\nsimple.txt:37:1: syntax error: non-declaration statement outside function body\nnested/file.go:39:1: syntax error: imports must appear before other declarations\n</error>\n",
			expected: []utils.FileInfo{
				{
					Original: "simple.txt",
					Resolved: simpleAbsPath,
				},
				{
					Original: "nested/file.go",
					Resolved: nestedAbsPath,
				},
			},
		},
		{
			name:    "Multiple unquoted path in search directories",
			content: "simple.txt:39:1: syntax error: imports must appear before other declarations",
			expected: []utils.FileInfo{
				{
					Original: "simple.txt",
					Resolved: simpleAbsPath,
				},
			},
		},
		{
			name:    "Double quoted path",
			content: "Check this file: \"" + filepath.Join(tempDir, "with space.txt") + "\"",
			expected: []utils.FileInfo{
				{
					Original: filepath.Join(tempDir, "with space.txt"),
					Resolved: spaceAbsPath,
				},
			},
		},
		{
			name:    "Single quoted path",
			content: "Check this file: '" + filepath.Join(tempDir, "nested/file.go") + "'",
			expected: []utils.FileInfo{
				{
					Original: filepath.Join(tempDir, "nested/file.go"),
					Resolved: nestedAbsPath,
				},
			},
		},
		{
			name:    "Backtick quoted path",
			content: "Check this file: `" + filepath.Join(tempDir, "nested/deeper/config.yaml") + "` here",
			expected: []utils.FileInfo{
				{
					Original: filepath.Join(tempDir, "nested/deeper/config.yaml"),
					Resolved: deeperAbsPath,
				},
			},
		},
		{
			name:    "Test files are found",
			content: "" + filepath.Join(tempDir, "nested/deeper/config.yaml") + ":37:1: syntax error: non-declaration statement outside function body",
			expected: []utils.FileInfo{
				{
					Original: filepath.Join(tempDir, "nested/deeper/config.yaml"),
					Resolved: deeperAbsPath,
				},
			},
		},
		{
			name: "Multiple paths",
			content: "Files: " +
				filepath.Join(tempDir, "simple.txt") + " and \"" +
				filepath.Join(tempDir, "with space.txt") + "\" and '" +
				filepath.Join(tempDir, "nested/file.go") + "'",
			expected: []utils.FileInfo{
				{
					Original: filepath.Join(tempDir, "simple.txt"),
					Resolved: simpleAbsPath,
				},
				{
					Original: filepath.Join(tempDir, "with space.txt"),
					Resolved: spaceAbsPath,
				},
				{
					Original: filepath.Join(tempDir, "nested/file.go"),
					Resolved: nestedAbsPath,
				},
			},
		},
		{
			name:     "Nonexistent files",
			content:  "This file doesn't exist: " + filepath.Join(tempDir, "nonexistent.txt"),
			expected: []utils.FileInfo{},
		},
		{
			name: "Mixed existing and nonexistent",
			content: "Files: " +
				filepath.Join(tempDir, "simple.txt") + " and " +
				filepath.Join(tempDir, "nonexistent.txt"),
			expected: []utils.FileInfo{
				{
					Original: filepath.Join(tempDir, "simple.txt"),
					Resolved: simpleAbsPath,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := finder.FindFilesInText(tt.content)

			assert.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(results), "Number of found files should match expected")

			// Create maps for easier comparison
			expectedMap := make(map[string]utils.FileInfo)
			for _, e := range tt.expected {
				expectedMap[e.Resolved] = e
			}

			resultsMap := make(map[string]utils.FileInfo)
			for _, r := range results {
				resultsMap[r.Resolved] = r
			}

			// Compare maps
			assert.Equal(t, expectedMap, resultsMap)
		})
	}
}

func TestExtractCandidates(t *testing.T) {
	finder := utils.NewFileFinder([]string{".", "/tmp"})

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "Empty string",
			content:  "",
			expected: []string{},
		},
		{
			name:     "Simple words",
			content:  "word1 word2 word3",
			expected: []string{"word1", "word2", "word3"},
		},
		{
			name:     "Double quotes",
			content:  "before \"quoted text\" after",
			expected: []string{"after", "before", "quoted text"},
		},
		{
			name:     "Single quotes",
			content:  "before 'quoted text' after",
			expected: []string{"after", "before", "quoted text"},
		},
		{
			name:     "Backticks",
			content:  "before `quoted text` after",
			expected: []string{"after", "before", "quoted text"},
		},
		{
			name:     "Mixed quotes",
			content:  "\"double\" 'single' `backtick`",
			expected: []string{"backtick", "double", "single"},
		},
		{
			name:     "Nested quotes not supported",
			content:  "\"outer 'inner' text\"",
			expected: []string{"outer 'inner' text"},
		},
		{
			name:     "Unclosed quotes",
			content:  "before \"unclosed",
			expected: []string{"before"},
		},
		{
			name:     "Path-like strings",
			content:  "/path/to/file ./relative/path ../parent/path",
			expected: []string{"../parent/path", "./relative/path", "/path/to/file"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := finder.ExtractCandidates(tt.content)
			assert.Equal(t, tt.expected, results)
		})
	}
}

func TestResolvePath(t *testing.T) {
	finder := utils.NewFileFinder([]string{})

	// Create a temp file for testing
	tempFile, err := os.CreateTemp("", "test-file-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	tempPath := tempFile.Name()
	tempAbsPath, _ := filepath.Abs(tempPath)

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{
			name:      "Absolute path",
			path:      tempAbsPath,
			wantError: false,
		},
		{
			name:      "Relative path",
			path:      filepath.Base(tempPath),
			wantError: false,
		},
		{
			name:      "Home directory",
			path:      "~/somefile",
			wantError: false, // This won't error, but the file won't exist
		},
		{
			name:      "Empty path",
			path:      "",
			wantError: false, // This will resolve to current directory
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolved, err := finder.ResolvePath(tt.path)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resolved)
				assert.True(t, filepath.IsAbs(resolved), "Resolved path should be absolute")
			}
		})
	}
}

func TestPathExists(t *testing.T) {
	finder := utils.NewFileFinder([]string{})

	// Create a temp file for testing
	tempFile, err := os.CreateTemp("", "test-file-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "Existing file",
			path:     tempFile.Name(),
			expected: true,
		},
		{
			name:     "Non-existent file",
			path:     tempFile.Name() + ".nonexistent",
			expected: false,
		},
		{
			name:     "Empty path",
			path:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists := finder.PathExists(tt.path)
			assert.Equal(t, tt.expected, exists)
		})
	}
}
