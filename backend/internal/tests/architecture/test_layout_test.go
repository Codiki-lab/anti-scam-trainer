package architecture_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllGoTestsLiveInInternalTests(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	testsRoot := filepath.Join(root, "internal", "tests")
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		if relative, err := filepath.Rel(testsRoot, path); err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			t.Errorf("Go test must live under internal/tests: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
