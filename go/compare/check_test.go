package compare

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestComment(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "checks")
	changed := Question("1-changed")
	changed.Status, changed.Result, changed.Scope = "✅", "Equal to the reference", "Example farm."
	drift := Question("2-drift")
	drift.Status, drift.Result, drift.Details = "❌", "Differs from gobees", "**Daily band:** 3 days."
	for _, c := range []Check{changed, drift} {
		if err := WriteCheck(dir, c); err != nil {
			t.Fatal(err)
		}
	}
	checks, err := ReadChecks(dir)
	if err != nil || len(checks) != 2 {
		t.Fatalf("read %d checks: %v", len(checks), err)
	}

	// netlogo and speed did not write a check: they failed to give a result.
	comment := Comment(checks, "https://example.org/run")
	for _, want := range []string{
		CommentMarker,
		"| ✅ | Did this step change the results? | Equal to the reference |",
		"| ❌ | Has the model drifted away from gobees? | Differs from gobees |",
		"| ❌ | Does it still match NetLogo? | no result, see the [run](https://example.org/run) |",
		"| ❌ | How fast is it? | no result",
		"- **Did this step change the results?** Example farm.",
		"<summary>Has the model drifted away from gobees?</summary>\n\n**Daily band:** 3 days.",
	} {
		if !strings.Contains(comment, want) {
			t.Errorf("comment lacks %q:\n%s", want, comment)
		}
	}
	if strings.Count(comment, "<details>") != 2 {
		t.Errorf("want the runs and the details of drift only:\n%s", comment)
	}
}

func TestReadChecksWithoutFolder(t *testing.T) {
	checks, err := ReadChecks(filepath.Join(t.TempDir(), "missing"))
	if err != nil || len(checks) != 0 {
		t.Errorf("got %v, %v", checks, err)
	}
}
