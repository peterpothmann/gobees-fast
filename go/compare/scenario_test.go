package compare

import (
	"strings"
	"testing"

	"gobees/input"
	"gobees/params"
)

func TestScenarioParameterFiles(t *testing.T) {
	files := map[string]params.ParameterFile{}
	for _, s := range Scenarios {
		text, err := s.ParameterFile()
		if err != nil {
			t.Fatal(err)
		}
		if files[s.Name], err = input.ParseParameterFile(strings.NewReader(text)); err != nil {
			t.Fatalf("%s: %v", s.Name, err)
		}
		if files[s.Name].StopExtinct {
			t.Errorf("%s: StopExtinct? must be false", s.Name)
		}
	}
	if got := files["step7_badgers"].N_Badgers; got != 2 {
		t.Errorf("badgers: N_Badgers = %v, want 2", got)
	}
	want := "_SYSTEM_Example_Farm_Foodsources.txt Example_Farm_PlotTest_Foodsources.txt"
	if got := files["step7_crop"].CropRotationListAsString; got != want {
		t.Errorf("crop rotation: CropRotationListAsString = %q, want %q", got, want)
	}
	if got := files["step7"].CropRotationListAsString; got != "" {
		t.Errorf("example farm: CropRotationListAsString = %q, want none", got)
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
