package compare

import (
	"path/filepath"
	"slices"
	"testing"

	"gobees/input"
	"gobees/params"
)

func TestScenarioRunFiles(t *testing.T) {
	// The run files of ScenarioFiles, as gobees-fast reads them. tools/netlogo2gobees checks that they are the
	// conversion of the parameter files (ParameterFile).
	dir := t.TempDir()
	if err := WriteInputFiles(dir); err != nil {
		t.Fatal(err)
	}
	runs := map[string]params.Params{}
	for _, s := range Scenarios {
		p, err := input.Load(filepath.Join(dir, scenariosDir, s.ConfigFile()))
		if err != nil {
			t.Fatalf("%s: %v", s.Name, err)
		}
		if p.File.StopExtinct {
			t.Errorf("%s: StopExtinct? must be false", s.Name)
		}
		runs[s.Name] = p
	}
	if got := runs["step7_badgers"].File.N_Badgers; got != 2 {
		t.Errorf("badgers: N_Badgers = %v, want 2", got)
	}
	want := []string{"_SYSTEM_Example_Farm_Foodsources.yaml", "Example_Farm_PlotTest_Foodsources.yaml"}
	if got := runs["step7_crop"].Input.Derived.CropRotationList; !slices.Equal(got, want) {
		t.Errorf("crop rotation: CropRotationList = %q, want %q", got, want)
	}
	if got := runs["step7"].Input.Derived.CropRotationList; len(got) != 0 {
		t.Errorf("example farm: CropRotationList = %q, want none", got)
	}
	if got := len(runs["step8_species"].Input.Species); got != 7 {
		t.Errorf("several species: %d species, want 7", got)
	}
}

func TestSetParameter(t *testing.T) {
	text := "A,\"B \",C\r\n1,2,3\r\n"
	got, err := setParameter(text, "B", "7")
	if err != nil || got != "A,\"B \",C\r\n1,7,3\r\n" {
		t.Errorf("got %q, %v", got, err)
	}
	if _, err := setParameter(text, "D", "1"); err == nil {
		t.Error("an unknown parameter must be an error")
	}
}
