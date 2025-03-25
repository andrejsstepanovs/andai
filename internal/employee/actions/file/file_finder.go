package file

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type Finder struct {
	SearchDirectories []string
}

type Info struct {
	Original string // Original path as found in text
	Resolved string // Absolute filesystem path
}

type Infos []Info

func (fi *Infos) GetAbsolutePaths() []string {
	paths := make([]string, 0, len(*fi))
	for _, f := range *fi {
		paths = append(paths, f.Resolved)
	}
	return paths
}

func (fi *Infos) String() string {
	return strings.Join(fi.GetAbsolutePaths(), ", ")
}

func NewFileFinder(searchDirectories []string) *Finder {
	return &Finder{
		SearchDirectories: searchDirectories,
	}
}

// FindFilesInText extracts valid filesystem paths from text content
func (f *Finder) FindFilesInText(content string) (Infos, error) {
	// Step 1: Split content into potential path candidates
	candidates := f.ExtractFilenameCandidates(content)

	// Step 2: Validate each candidate against the filesystem
	var results []Info
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
func (f *Finder) tryFindInDirectories(filename string, seen map[string]struct{}, results *[]Info, originalPath string) {
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
			*results = append(*results, Info{
				Original: originalPath,
				Resolved: absPath,
			})
		}
	}
}

func (f *Finder) ExtractFilenameCandidates(content string) []string {
	lines := strings.Split(content, "\n")
	var candidates []string
	for _, line := range lines {
		candidates = append(candidates, f.ExtractFilenameCandidatesLine(line)...)
	}
	return candidates
}

// ExtractFilenameCandidates splits a line of text into potential path candidates.
// It handles quoted strings (with double quotes, single quotes, or backticks)
// and extracts potential filesystem paths.
func (f *Finder) handleQuote(char rune, current *strings.Builder, inQuote *bool, quoteChar *rune, candidates *[]string) bool {
	if *inQuote && *quoteChar == char {
		// End of quote
		*candidates = append(*candidates, current.String())
		current.Reset()
		*inQuote = false
		return true
	} else if !*inQuote {
		// Start of quote
		*inQuote = true
		*quoteChar = char
		current.Reset()
		return true
	}
	return false
}

// processWord adds a word to candidates if it's not empty
func (f *Finder) processWord(word string, candidates *[]string) {
	if word != "" {
		*candidates = append(*candidates, word)
	}
}

// processCandidate handles splitting and processing of a single candidate
func (f *Finder) processCandidate(candidate string) []string {
	var processed []string

	// Split by colon for line/column numbers
	parts := strings.Split(candidate, ":")
	for _, part := range parts {
		if part != "" {
			processed = append(processed, part)
		}
	}
	// Keep original candidate too
	processed = append(processed, candidate)
	return processed
}

// uniqueAndSort deduplicates and sorts candidates
func (f *Finder) uniqueAndSort(candidates []string) []string {
	unique := make(map[string]struct{})
	for _, candidate := range candidates {
		unique[candidate] = struct{}{}
	}

	result := make([]string, 0, len(unique))
	for candidate := range unique {
		result = append(result, candidate)
	}
	sort.Strings(result)
	return result
}

func (f *Finder) ExtractFilenameCandidatesLine(content string) []string {
	var candidates []string
	var current strings.Builder
	var inQuote bool
	var quoteChar rune

	// Process content character by character
	for _, char := range content {
		switch {
		case char == '"' || char == '\'' || char == '`':
			if !f.handleQuote(char, &current, &inQuote, &quoteChar, &candidates) {
				current.WriteRune(char)
			}
		case unicode.IsSpace(char) && !inQuote:
			f.processWord(current.String(), &candidates)
			current.Reset()
		default:
			current.WriteRune(char)
		}
	}

	// Handle last word if not in quotes
	if current.Len() > 0 && !inQuote {
		f.processWord(current.String(), &candidates)
	}

	// Process all candidates
	var processedCandidates []string
	for _, candidate := range candidates {
		processedCandidates = append(processedCandidates, f.processCandidate(candidate)...)
	}

	return f.uniqueAndSort(processedCandidates)
}

// ResolvePath attempts to convert a path string to an absolute path
func (f *Finder) ResolvePath(candidate string) (string, error) {
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
func (f *Finder) PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDirectory checks if a path is a directory
func (f *Finder) IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
