package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"gobees/input"
)

// writeSites writes into dir two sites and three run files: sites.yaml (the example farm with the variation sites
// sites/*/site.yaml), a.yaml and b.yaml (the example farm with the landscape files of site A and of site B). Site A
// has the landscape of the example farm, site B the same start and a crop rotation into the landscape of the plot
// test. All run 370 days, over the turn of the year.
func writeSites(t *testing.T, dir string) {
	t.Helper()
	cfg, landscapes := input.ExampleFarmConfig()
	farm := landscapes[cfg.Landscape]
	open := func(name string) (io.ReadCloser, error) { return os.Open(name) }
	plot, _, _, err := input.ReadLandscape(filepath.Join("compare", "scenarios", "Example_Farm_PlotTest_Foodsources.yaml"),
		open)
	if err != nil {
		t.Fatal(err)
	}
	days := 370
	cfg.Run = &input.Run{Seed: 2, Days: &days, Repetitions: 2, Threads: 2}

	a, b := cfg, cfg
	a.Name, a.Landscape = "a", "sites/A/start.yaml"
	b.Name, b.Landscape = "b", "sites/B/start.yaml"
	b.CropRotation = []string{"sites/B/start.yaml", "sites/B/year2.yaml"}
	writeRunFile(t, dir, "a.yaml", a, map[string]input.Landscape{a.Landscape: farm})
	writeRunFile(t, dir, "b.yaml", b, map[string]input.Landscape{b.Landscape: farm, "sites/B/year2.yaml": plot})

	s := cfg
	s.Name, s.Landscape = "sites", ""
	s.Variations = []input.Variation{{Sites: input.Patterns{"sites/*/site.yaml"}}}
	writeRunFile(t, dir, "sites.yaml", s, nil)
	for name, text := range map[string]string{
		"sites/A/site.yaml": "landscape: start.yaml\n",
		"sites/B/site.yaml": "landscape: start.yaml\ncropRotation: [start.yaml, year2.yaml]\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSitesExperiment(t *testing.T) {
	// The runs of a site are the runs of a run file with its landscape files: the same rows with the same seeds. The
	// column site after run and seed holds the name of the site. The provenance lists the site files it ran, with the
	// SHA-256 of their landscape files, and repeats the runs byte for byte.
	dir := t.TempDir()
	writeSites(t, dir)
	out := filepath.Join(dir, "out")
	for _, name := range []string{"sites", "a", "b"} {
		c := parse(t, "-c", filepath.Join(dir, name+".yaml"), "-o", out, "--columns", "colonies,workers")
		if err := c.run(&bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
	}
	daily, prov := outputOf(out, "sites")
	sites := readFile(t, daily)
	header := strings.SplitN(string(sites), "\n", 2)[0]
	if !strings.HasPrefix(header, "run\tseed\tsite\tt\t") {
		t.Errorf("header %q, want run, seed, site, t", header)
	}
	aDaily, _ := outputOf(out, "a")
	bDaily, _ := outputOf(out, "b")
	want := append(dropFields(t, readFile(t, aDaily), 2), dropFields(t, readFile(t, bDaily), 2)...)
	if got := dropFields(t, sites, 3); !slices.Equal(got, want) {
		t.Errorf("the rows of the sites differ from those of a.yaml and b.yaml (%d and %d rows)", len(got),
			len(want))
	}
	var names []string
	for _, l := range strings.Split(strings.TrimSpace(string(sites)), "\n")[1:] {
		names = append(names, strings.Split(l, "\t")[2])
	}
	if names[0] != "A" || names[len(names)-1] != "B" {
		t.Errorf("column site from %q to %q, want A to B", names[0], names[len(names)-1])
	}

	info, err := input.Read(prov)
	if err != nil {
		t.Fatalf("the provenance is no valid run file: %v", err)
	}
	records := info.Config.Run.Sites
	if len(records) != 2 || records[0].Site != "A" || len(records[0].Landscapes) != 1 ||
		len(records[1].Landscapes) != 2 || records[1].Landscapes[1].Source != "sites/B/year2.yaml" ||
		len(records[1].Landscapes[1].SHA256) != 64 || len(info.Config.Run.Landscapes) != 0 {
		t.Errorf("run.sites = %+v, want A with 1 and B with 2 landscape files, no copies", records)
	}
	if files := info.Config.Variations[0].Sites; len(files) != 2 || files[0] != "../../sites/A/site.yaml" {
		t.Errorf("sites of the provenance = %q, want the site files relative to the provenance", files)
	}
	again := filepath.Join(dir, "again")
	c := parse(t, "-c", prov, "-o", again)
	if err := c.run(&bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	againDaily, _ := outputOf(again, "sites")
	if !bytes.Equal(readFile(t, againDaily), sites) {
		t.Error("the runs repeated from the provenance have a different daily output")
	}
}

func TestSitesChangedLandscape(t *testing.T) {
	// a landscape file that changes after the start is an error of the runs of its combination
	dir := t.TempDir()
	writeSites(t, dir)
	c := parse(t, "-c", filepath.Join(dir, "sites.yaml"))
	e, err := c.newExperiment(c.Config[0])
	if err != nil {
		t.Fatal(err)
	}
	year2 := filepath.Join(dir, "sites", "B", "year2.yaml")
	if err := os.WriteFile(year2, append(readFile(t, year2), "# changed\n"...), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := e.params(1); err == nil || !strings.Contains(err.Error(), "changed after the start") {
		t.Errorf("error = %v, want the changed landscape file", err)
	}
}
