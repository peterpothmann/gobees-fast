package main

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// TestGeneratedFilesUpToDate checks that the run files in the repository are the conversion of the NetLogo files
// (Generate). After a change of the NetLogo files, the scenarios or the run file format: go run . generate
func TestGeneratedFilesUpToDate(t *testing.T) {
	files, _, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join("..", "..")
	for name, want := range files {
		got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			t.Errorf("%s: %v (run: go run . generate)", name, err)
			continue
		}
		// git may check the files out with CRLF line endings
		if !bytes.Equal(bytes.ReplaceAll(got, []byte("\r\n"), []byte("\n")), want) {
			t.Errorf("%s is not up to date (run: go run . generate)", name)
		}
	}
	// no other files in the folders of the generated files
	for _, dir := range []string{exampleFarmFolder, scenariosFolder} {
		entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(dir)))
		if err != nil {
			t.Fatal(err)
		}
		extra := []string{}
		for _, e := range entries {
			if _, ok := files[dir+"/"+e.Name()]; !ok {
				extra = append(extra, e.Name())
			}
		}
		sort.Strings(extra)
		if len(extra) > 0 {
			t.Errorf("%s: files that Generate does not write: %s", dir, strings.Join(extra, ", "))
		}
	}
}
