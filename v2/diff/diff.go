package diff

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

type Diff struct {
	lines []line
}

// AgainstFile is a testing helper function that takes the got string and want file
// and diffs the got string against the contents of the want file.
// if got and want are the same an empty string is returned
func AgainstFile(t *testing.T, got, wantFilePath string) string {
	t.Helper()

	wantBytes, err := os.ReadFile(wantFilePath)
	if err != nil {
		t.Fatalf("Failed to read diff file %s", wantFilePath)
	}

	want := string(wantBytes)
	if got == want {
		// there were no diffs so return an empty string
		return ""
	}

	diff := NewDiff(got, want)
	return PrettyPrint(diff)
}

// UpdateFile updates the want file to match the got file contents.
func UpdateFile(t *testing.T, got, wantFilePath string) {
	t.Helper()

	err := os.WriteFile(wantFilePath, []byte(got), 0o0655)
	if err != nil {
		t.Fatalf("failed to write diff file %s", wantFilePath)
	}

	t.Errorf("File Successfully updated, please remove UpdateFile()")
}

// lineStatus is the status of the line diff
type lineStatus int

const (
	match = lineStatus(iota)
	added
	removed
	modified
)

// line is a line in the diff
type line struct {
	number  int
	content string
	status  lineStatus
}

// toLines converts a string into diff lines
func toLines(s string) []line {
	splitLines := strings.Split(s, "\n")
	lines := []line{}
	for i, l := range splitLines {
		lines = append(lines, line{number: i + 1, content: l})
	}

	return lines
}

// NewDiff creates a new diff from two strings
func NewDiff(got, want string) Diff {
	gotLines := toLines(got)
	wantLines := toLines(want)

	offset := 0
	combined := []line{}
	for i := range gotLines {
		a := gotLines[i]

		if i+offset >= len(wantLines) {
			a.status = added
			combined = append(combined, a)
			continue
		}

		b := wantLines[i+offset]
		if a == b {
			combined = append(combined, a)
			continue
		}

		// look for the got line in the remaining wantLines
		// if we can find it, then the lines we skipped were removed
		found := findMatchingLine(a, wantLines[i+offset:])
		for j := 0; j < found; j++ {
			b := wantLines[i+j+offset]
			b.status = removed
			combined = append(combined, b)
		}
		if found != -1 {
			combined = append(combined, a)
			offset += found
			continue
		}

		// look for the want line in the remaining gotLines
		// if we can find it, then the current line is a newly added
		found = findMatchingLine(b, gotLines[i:])
		if found != -1 {
			a.status = added
			combined = append(combined, a)
			offset -= 1
			continue
		}

		// if the line isn't added or deleted, it's modified
		a.status = modified
		combined = append(combined, a)
		b.status = modified
		combined = append(combined, b)
	}

	for i := offset + len(gotLines); i < len(wantLines); i++ {
		b := wantLines[i]
		b.status = removed
		combined = append(combined, b)
	}

	return Diff{lines: combined}
}

// findMatchingLine searches in search to find want. If it can't be found it returns -1
func findMatchingLine(want line, search []line) int {
	for i, s := range search {
		if want.content == s.content {
			return i
		}
	}

	return -1
}

// PrettyPrint converts the diff into a colored string
func PrettyPrint(diff Diff) string {
	diffs := []string{}
	for _, line := range diff.lines {
		s := fmt.Sprintf("%03d    %s", line.number, line.content)
		switch line.status {
		case match:
			s = colorLine(s, 250, 235)
		case added:
			s = colorLine(s, 28, 22)
		case removed:
			s = colorLine(s, 160, 52)
		case modified:
			s = colorLine(s, 226, 235)
		}

		diffs = append(diffs, s)
	}

	return strings.Join(diffs, "\n")
}

// colorLine colors the line with primaryColors for regular characters and
// the secondaryColor for space characters
// https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797
func colorLine(line string, primaryColor, secondaryColor int) string {
	newLine := ""
	color := false
	for _, r := range line {
		switch r {
		case ' ':
			if color {
				newLine += fmt.Sprintf("\033[0m\033[38;5;%dm", secondaryColor)
			}
			color = false
			newLine += "·"
		case '\t':
			if color {
				newLine += fmt.Sprintf("\033[0m\033[38;5;%dm", secondaryColor)
			}
			color = false
			newLine += "■■■■"
		default:
			if !color {
				newLine += fmt.Sprintf("\033[0m\033[38;5;%dm", primaryColor)
			}
			color = true
			newLine += string(r)
		}
	}

	return newLine[4:] + "\033[0m"
}
