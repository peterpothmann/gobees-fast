package input

import (
	"io"
	"slices"
	"strings"
	"testing"
)

// exampleFarmWith returns the run file of the example farm with the text extra appended (a section at the end), read
// with its landscape file.
func exampleFarmWith(t *testing.T, extra string) (Files, error) {
	t.Helper()
	r, err := openExampleFarm(ExampleFarmFile)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	data = append([]byte(strings.ReplaceAll(string(data), "\r\n", "\n")), extra...)
	return readFrom(data, openExampleFarm, Files{})
}

func TestCombinations(t *testing.T) {
	// full factorial, the first variation changes slowest. Each combination is the run file with these values and
	// without the section variations, the landscape files are those of the run file.
	f, err := exampleFarmWith(t, `variations:
  - key: parameters.badgers
    values: [0, 2]
  - key: constants.foraging.maxForagingRange
    values: [500, 758.5]
  - key: initialQueens.0.queens
    values: [100]
`)
	if err != nil {
		t.Fatal(err)
	}
	combinations, err := f.Combinations()
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		values  []string
		badgers int
		rng     float64
		queens  int
	}{
		{[]string{"0", "500", "100"}, 0, 500, 100},
		{[]string{"0", "758.5", "100"}, 0, 758.5, 100},
		{[]string{"2", "500", "100"}, 2, 500, 100},
		{[]string{"2", "758.5", "100"}, 2, 758.5, 100},
	}
	if len(combinations) != len(want) {
		t.Fatalf("%d combinations, want %d", len(combinations), len(want))
	}
	for i, c := range combinations {
		w := want[i]
		if !slices.Equal(c.Values, w.values) {
			t.Errorf("combination %d: values %v, want %v", i, c.Values, w.values)
		}
		cfg := c.Files.Config
		if cfg.Parameters.Badgers != w.badgers || cfg.Constants.Foraging.MaxForagingRange != w.rng ||
			cfg.InitialQueens[0].Queens != w.queens {
			t.Errorf("combination %d: badgers %d, range %v, queens %v, want %d, %v, %d", i, cfg.Parameters.Badgers,
				cfg.Constants.Foraging.MaxForagingRange, cfg.InitialQueens[0].Queens, w.badgers, w.rng, w.queens)
		}
		if len(cfg.Variations) != 0 {
			t.Errorf("combination %d still has variations", i)
		}
		if _, ok := c.Files.Landscapes[cfg.Landscape]; !ok {
			t.Errorf("combination %d: landscape %s not read", i, cfg.Landscape)
		}
		// the other keys as in the run file
		other := cfg.Parameters
		other.Badgers = f.Config.Parameters.Badgers
		if other != f.Config.Parameters || cfg.Run.Seed != f.Config.Run.Seed {
			t.Errorf("combination %d: other keys changed", i)
		}
	}
	if got := f.Describe(combinations[3]); got != "parameters.badgers=2, constants.foraging.maxForagingRange=758.5, initialQueens.0.queens=100" {
		t.Errorf("Describe = %q", got)
	}
	// the run file itself is unchanged
	if f.Config.Parameters.Badgers != 0 || len(f.Config.Variations) != 3 {
		t.Errorf("the run file changed: badgers %d, %d variations", f.Config.Parameters.Badgers, len(f.Config.Variations))
	}
}

func TestCombinationsWithoutVariations(t *testing.T) {
	f := ExampleFarmFiles()
	combinations, err := f.Combinations()
	if err != nil {
		t.Fatal(err)
	}
	if len(combinations) != 1 || len(combinations[0].Values) != 0 || combinations[0].Files.Config.Parameters != f.Config.Parameters {
		t.Errorf("want one combination, the run file itself, got %+v", combinations)
	}
}

func TestCombinationNumber(t *testing.T) {
	// Combination(k) builds the combination k alone, the same as item k of Combinations. A number outside the
	// combinations is an error.
	f, err := exampleFarmWith(t, `variations:
  - key: parameters.badgers
    values: [0, 2, 4]
  - key: parameters.sexLocus
    values: [false, true]
`)
	if err != nil {
		t.Fatal(err)
	}
	if n := f.CombinationCount(); n != 6 {
		t.Fatalf("CombinationCount = %d, want 6", n)
	}
	all, err := f.Combinations()
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []int{5, 0, 3} {
		c, err := f.Combination(k)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(c.Values, all[k].Values) || c.Files.Config.Parameters != all[k].Files.Config.Parameters {
			t.Errorf("combination %d: values %q, want %q", k, c.Values, all[k].Values)
		}
	}
	for _, k := range []int{-1, 6} {
		if _, err := f.Combination(k); err == nil || !strings.Contains(err.Error(), "want 0 to 5") {
			t.Errorf("combination %d: error %v, want 0 to 5", k, err)
		}
	}
	if n := ExampleFarmFiles().CombinationCount(); n != 1 {
		t.Errorf("CombinationCount without variations = %d, want 1", n)
	}
}

func TestCombinationsListValue(t *testing.T) {
	// a value can be a list, its text in the output is one line
	f, err := exampleFarmWith(t, `variations:
  - key: stewardshipOptions
    values: [[], [Plot, Margin]]
`)
	if err != nil {
		t.Fatal(err)
	}
	combinations, err := f.Combinations()
	if err != nil {
		t.Fatal(err)
	}
	if combinations[1].Values[0] != "[Plot, Margin]" || !slices.Equal(combinations[1].Files.Config.StewardshipOptions,
		[]string{"Plot", "Margin"}) || len(combinations[0].Files.Config.StewardshipOptions) != 0 {
		t.Errorf("values %q and %q, options %v", combinations[0].Values[0], combinations[1].Values[0],
			combinations[1].Files.Config.StewardshipOptions)
	}
}

func TestVariationErrors(t *testing.T) {
	for _, tc := range []struct {
		variations string
		want       string // in the error
	}{
		{"  - key: run.seed\n    values: [1, 2]\n", "run.seed cannot vary"},
		{"  - key: name\n    values: [a, b]\n", "name cannot vary"},
		{"  - key: output.columns\n    values: [[eggs]]\n", "output.columns cannot vary"},
		{"  - key: landscape\n    values: [a.yaml]\n", "landscape cannot vary"},
		{"  - key: cropRotation\n    values: [[a.yaml, b.yaml]]\n", "cropRotation cannot vary"},
		{"  - key: parameters.badgers\n    values: [1]\n  - key: parameters.badgers\n    values: [2]\n", "listed twice"},
		{"  - key: parameters.badgers\n    values: []\n", "no values"},
		{"  - values: [1]\n", "key is missing"},
	} {
		_, err := exampleFarmWith(t, "variations:\n"+tc.variations)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%q: error = %v, want %q", tc.variations, err, tc.want)
		}
	}

	// errors of a combination: the strict reading of the run file finds them, the message names the combination
	for _, tc := range []struct {
		variations string
		want       string
	}{
		{"  - key: parameters.nonsense\n    values: [1]\n", "parameters.nonsense=1"},
		{"  - key: parameters.badgers\n    values: [2.5]\n", "not a whole number"},
		{"  - key: initialQueens.5.queens\n    values: [1]\n", "item number from 0 to"},
		{"  - key: parameters.badgers.x\n    values: [1]\n", "is a value"},
	} {
		f, err := exampleFarmWith(t, "variations:\n"+tc.variations)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Combinations(); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%q: error = %v, want %q", tc.variations, err, tc.want)
		}
	}
}
