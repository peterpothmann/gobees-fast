package input

import (
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// writeFile writes text into the file name in dir, with missing folders.
func writeFile(t *testing.T, dir, name, text string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// exampleFarmText returns the file name of the example farm as text, with \n line ends.
func exampleFarmText(t *testing.T, name string) string {
	t.Helper()
	r, err := openExampleFarm(name)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n")
}

// writeSites writes into a new folder the run file run.yaml (the example farm without its key landscape, with the
// text variations appended) and two sites with the landscape of the example farm: sites/A with the start only,
// sites/B with a crop rotation of start.yaml and year2.yaml. Each site holds the tables of the example farm, which
// start.yaml and year2.yaml name. It returns the folder.
func writeSites(t *testing.T, variations string) string {
	t.Helper()
	dir := t.TempDir()
	var run []string
	for _, l := range strings.Split(exampleFarmText(t, ExampleFarmFile), "\n") {
		if !strings.HasPrefix(l, "landscape:") {
			run = append(run, l)
		}
	}
	writeFile(t, dir, "run.yaml", strings.Join(run, "\n")+variations)
	landscape := exampleFarmText(t, "_SYSTEM_Example_Farm_Foodsources.yaml")
	writeFile(t, dir, "sites/A/site.yaml", "landscape: start.yaml\n")
	writeFile(t, dir, "sites/A/start.yaml", landscape)
	writeFile(t, dir, "sites/B/site.yaml", "landscape: start.yaml\ncropRotation: [start.yaml, year2.yaml]\n")
	writeFile(t, dir, "sites/B/start.yaml", landscape)
	writeFile(t, dir, "sites/B/year2.yaml", landscape)
	for _, table := range []string{"_SYSTEM_Example_Farm_Foodsources_patches.csv",
		"_SYSTEM_Example_Farm_Foodsources_habitats.csv"} {
		text := exampleFarmText(t, table)
		writeFile(t, dir, "sites/A/"+table, text)
		writeFile(t, dir, "sites/B/"+table, text)
	}
	return dir
}

func TestSites(t *testing.T) {
	// each site file sets landscape and cropRotation, relative to its folder. The sites vary with the other
	// variations (full factorial), the value of a site is the name of its folder.
	dir := writeSites(t, `variations:
  - sites: sites/*/site.yaml
  - key: parameters.badgers
    values: [0, 2]
`)
	f, err := Read(filepath.Join(dir, "run.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Sites) != 2 || f.Sites[0].Name != "A" || f.Sites[1].Name != "B" || len(f.Landscapes) != 0 {
		t.Fatalf("sites %+v, %d landscapes, want A and B and no landscape of the run file", f.Sites,
			len(f.Landscapes))
	}
	if n := f.CombinationCount(); n != 4 {
		t.Fatalf("CombinationCount = %d, want 4", n)
	}
	want := []struct {
		values    []string
		landscape string
		rotation  []string
	}{
		{[]string{"A", "0"}, "sites/A/start.yaml", nil},
		{[]string{"A", "2"}, "sites/A/start.yaml", nil},
		{[]string{"B", "0"}, "sites/B/start.yaml", []string{"sites/B/start.yaml", "sites/B/year2.yaml"}},
		{[]string{"B", "2"}, "sites/B/start.yaml", []string{"sites/B/start.yaml", "sites/B/year2.yaml"}},
	}
	for k, w := range want {
		c, err := f.Combination(k)
		if err != nil {
			t.Fatal(err)
		}
		cfg := c.Files.Config
		if !slices.Equal(c.Values, w.values) || cfg.Landscape != w.landscape || !slices.Equal(cfg.CropRotation,
			w.rotation) || len(cfg.Variations) != 0 {
			t.Errorf("combination %d: values %q, landscape %q, crop rotation %q, want %q, %q, %q", k, c.Values,
				cfg.Landscape, cfg.CropRotation, w.values, w.landscape, w.rotation)
		}
		for _, name := range landscapeNames(cfg) {
			if _, ok := c.Files.Landscapes[name]; !ok || len(c.Files.Raw[name]) == 0 {
				t.Errorf("combination %d: landscape file %s not read", k, name)
			}
		}
	}
	c, err := f.Combination(3)
	if err != nil {
		t.Fatal(err)
	}
	if got := f.Describe(c); got != "site=B, parameters.badgers=2" {
		t.Errorf("Describe = %q", got)
	}
	if got := f.Config.Variations[0].Column(); got != "site" {
		t.Errorf("Column = %q, want site", got)
	}

	// a list of files: in this order, not sorted
	writeFile(t, dir, "list.yaml", strings.Replace(exampleFarmText(t, ExampleFarmFile), "landscape: _SYSTEM_Example_Farm_Foodsources.yaml\n",
		"", 1)+"variations:\n  - sites: [sites/B/site.yaml, sites/A/site.yaml]\n")
	f, err = Read(filepath.Join(dir, "list.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Sites) != 2 || f.Sites[0].Name != "B" || f.Sites[1].Name != "A" {
		t.Errorf("sites %+v, want B, A", f.Sites)
	}
}

func TestSiteErrors(t *testing.T) {
	// errors of the section variations and of the site files, found when the run file is read
	for _, tc := range []struct {
		variations string
		prepare    func(dir string)
		want       string // in the error
	}{
		{"  - sites: nowhere/*/site.yaml\n", nil, "matches no file"},
		{"  - sites: [sites/A/site.yaml, sites/*/site.yaml]\n", nil, "matched twice"},
		{"  - sites: sites/*/site.yaml\n    key: parameters.badgers\n", nil, "sites has no key and no values"},
		{"  - sites: sites/A/site.yaml\n  - sites: sites/B/site.yaml\n", nil, "sites is listed twice"},
		{"  - sites: sites/*/site.yaml\n", func(dir string) {
			writeFile(t, dir, "sites/A/site.yaml", "landscape: start.yaml\nbadgers: 2\n")
		}, "field badgers not found"},
		{"  - sites: sites/*/site.yaml\n", func(dir string) {
			writeFile(t, dir, "sites/A/site.yaml", "cropRotation: [start.yaml]\n")
		}, "landscape is missing"},
		{"  - sites: sites/*/*/site.yaml\n", func(dir string) {
			writeFile(t, dir, "sites/x/C/site.yaml", "landscape: start.yaml\n")
			writeFile(t, dir, "sites/y/C/site.yaml", "landscape: start.yaml\n")
		}, `have the same name "C"`},
		{"  - sites: sites/*/site.yaml\n", func(dir string) {
			writeFile(t, dir, "sites/my site/site.yaml", "landscape: start.yaml\n")
		}, "the folder is the name of the site"},
		{"  - sites: site.yaml\n", func(dir string) {
			writeFile(t, dir, "site.yaml", "landscape: start.yaml\n")
		}, "the folder is the name of the site"},
	} {
		dir := writeSites(t, "variations:\n"+tc.variations)
		if tc.prepare != nil {
			tc.prepare(dir)
		}
		_, err := Read(filepath.Join(dir, "run.yaml"))
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%q: error = %v, want %q", tc.variations, err, tc.want)
		}
	}

	// the run file of the sites has no landscape of its own
	dir := writeSites(t, "variations:\n  - sites: sites/*/site.yaml\n")
	writeFile(t, dir, "own.yaml", exampleFarmText(t, ExampleFarmFile)+"variations:\n  - sites: sites/*/site.yaml\n")
	if _, err := Read(filepath.Join(dir, "own.yaml")); err == nil || !strings.Contains(err.Error(),
		"remove them from the run file") {
		t.Errorf("run file with landscape and sites: error = %v", err)
	}

	// a missing landscape file of a site is an error of its combination, naming the file
	if err := os.Remove(filepath.Join(dir, "sites", "B", "year2.yaml")); err != nil {
		t.Fatal(err)
	}
	f, err := Read(filepath.Join(dir, "run.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Combination(0); err != nil {
		t.Errorf("site A: %v", err)
	}
	if _, err := f.Combination(1); err == nil || !strings.Contains(err.Error(), "year2.yaml") {
		t.Errorf("site B without year2.yaml: error = %v", err)
	}
}
