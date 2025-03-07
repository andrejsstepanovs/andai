package utils

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type FileFinder struct {
	SearchDirectories []string
}

type FileInfo struct {
	Original string // Original path as found in text
	Resolved string // Absolute filesystem path
}

type FileInfos []FileInfo

func (fi *FileInfos) GetAbsolutePaths() []string {
	paths := make([]string, 0, len(*fi))
	for _, f := range *fi {
		paths = append(paths, f.Resolved)
	}
	return paths
}

func (fi *FileInfos) String() string {
	return strings.Join(fi.GetAbsolutePaths(), ", ")
}

func NewFileFinder(searchDirectories []string) *FileFinder {
	return &FileFinder{
		SearchDirectories: searchDirectories,
	}
}

// FindFilesInText extracts valid filesystem paths from text content
func (f *FileFinder) FindFilesInText(content string) (FileInfos, error) {
	// Step 1: Split content into potential path candidates
	candidates := f.ExtractCandidates(content)

	// Step 2: Validate each candidate against the filesystem
	var results []FileInfo
	seen := make(map[string]struct{})

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}

		f.tryFindInDirectories(candidate, seen, &results, candidate)
	}

	return results, nil
}

// tryFindInDirectories attempts to find a file in all search directories
func (f *FileFinder) tryFindInDirectories(filename string, seen map[string]struct{}, results *[]FileInfo, originalPath string) {
	for _, dir := range f.SearchDirectories {
		if dir != "" && !f.PathExists(dir) {
			continue
		}

		searchPath := filepath.Join(dir, filename)
		absPath, err := f.ResolvePath(searchPath)
		if err != nil {
			continue
		}

		// Check if the path exists, is a file (not a directory), and has an extension
		if !f.PathExists(absPath) || f.IsDirectory(absPath) || filepath.Ext(absPath) == "" {
			continue
		}

		// Add to results if not seen before
		if _, exists := seen[absPath]; !exists {
			seen[absPath] = struct{}{}
			*results = append(*results, FileInfo{
				Original: originalPath,
				Resolved: absPath,
			})
		}
	}
}

func (f *FileFinder) ExtractCandidates(content string) []string {
	lines := strings.Split(content, "\n")
	var candidates []string
	for _, line := range lines {
		candidates = append(candidates, f.ExtractCandidatesLine(line)...)
	}
	return candidates
}

// ExtractCandidates splits text into potential path candidates
func (f *FileFinder) ExtractCandidatesLine(content string) []string {
	var candidates []string
	var current strings.Builder
	var inQuote, inSingleQuote, inBacktick bool
	var quoteChar rune

	// Process content character by character
	for _, char := range content {
		switch {
		case char == '"' && !inSingleQuote && !inBacktick:
			if inQuote && quoteChar == '"' {
				// End of double quote
				candidates = append(candidates, current.String())
				current.Reset()
				inQuote = false
			} else if !inQuote {
				// Start of double quote
				inQuote = true
				quoteChar = '"'
				current.Reset()
			} else {
				current.WriteRune(char)
			}

		case char == '\'' && !inQuote && !inBacktick:
			if inSingleQuote && quoteChar == '\'' {
				// End of single quote
				candidates = append(candidates, current.String())
				current.Reset()
				inSingleQuote = false
			} else if !inSingleQuote {
				// Start of single quote
				inSingleQuote = true
				quoteChar = '\''
				current.Reset()
			} else {
				current.WriteRune(char)
			}

		case char == '`' && !inQuote && !inSingleQuote:
			if inBacktick && quoteChar == '`' {
				// End of backtick
				candidates = append(candidates, current.String())
				current.Reset()
				inBacktick = false
			} else if !inBacktick {
				// Start of backtick
				inBacktick = true
				quoteChar = '`'
				current.Reset()
			} else {
				current.WriteRune(char)
			}

		case unicode.IsSpace(char) && !inQuote && !inSingleQuote && !inBacktick:
			// Space outside quotes means end of current word
			if current.Len() > 0 {
				candidates = append(candidates, current.String())
				current.Reset()
			}

		default:
			// Add character to current word
			current.WriteRune(char)
		}
	}

	// Don't forget the last word if not in quotes
	if current.Len() > 0 && !inQuote && !inSingleQuote && !inBacktick {
		candidates = append(candidates, current.String())
	}

	// Process candidates for paths with line numbers and error messages
	processedCandidates := make([]string, 0)
	for _, candidate := range candidates {
		// Check if the candidate looks like a file path with line/column numbers
		parts := strings.Split(candidate, ":")
		for _, part := range parts {
			if part != "" {
				processedCandidates = append(processedCandidates, part)
			}
		}
		// Also keep the original candidate
		processedCandidates = append(processedCandidates, candidate)
	}

	uniquedCandidates := make(map[string]struct{})
	for _, candidate := range processedCandidates {
		uniquedCandidates[candidate] = struct{}{}
	}

	response := make([]string, 0, len(uniquedCandidates))
	for candidate := range uniquedCandidates {
		response = append(response, candidate)
	}
	// sort alphabetically
	sort.Strings(response)

	return response
}

// ResolvePath attempts to convert a path string to an absolute path
func (f *FileFinder) ResolvePath(candidate string) (string, error) {
	// Expand home directory if present
	if strings.HasPrefix(candidate, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		candidate = filepath.Join(home, strings.TrimPrefix(candidate, "~"))
	}

	// Clean and make absolute
	cleanPath := filepath.Clean(candidate)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

// PathExists checks if a path exists in the filesystem
func (f *FileFinder) PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if a path is a directory
func (f *FileFinder) IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
