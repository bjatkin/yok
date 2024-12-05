package diff

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

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

	return Lines(got, want)
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

// Lines takes got and want, splits them into lines and combines them into
// a nice diff using colors and line numbers
func Lines(got, want string) string {
	gotLines := strings.Split(got, "\n")
	wantLines := strings.Split(want, "\n")
	lines := max(len(gotLines), len(wantLines))

	diff := []string{}
	for i := 0; i < lines; i++ {
		a := "\033[3m[empty\033[0m \033[3mline]\033[0m"
		b := "\033[3m[empty\033[0m \033[3mline]\033[0m"
		if len(gotLines) > i {
			a = gotLines[i]
		}
		if len(wantLines) > i {
			b = wantLines[i]
		}

		a = strings.ReplaceAll(a, " ", "\033[90m·\033[0m")
		b = strings.ReplaceAll(b, " ", "\033[90m·\033[0m")
		a = strings.ReplaceAll(a, "\t", "\033[90m↦\033[0m")
		b = strings.ReplaceAll(b, "\t", "\033[90m↦\033[0m")

		if a == b {
			a = fmt.Sprintf("\033[90m%3d\033[0m %s", i+1, a)
			diff = append(diff, a)
			continue
		}

		a = fmt.Sprintf("\033[90m%3d    \033[0m \033[31;1m got\033[0m %s", i+1, a)
		diff = append(diff, a)
		b = fmt.Sprintf("\033[90m    %3d\033[0m \033[32;1mwant\033[0m %s", i+1, b)
		diff = append(diff, "\033[32m"+b+"\033[0m")
	}

	return strings.Join(diff, "\n")
}
