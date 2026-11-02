package compare

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Check is the answer to one question of the check of a step: a row of the pull request comment and, if Details is
// not empty, a collapsed section below it (Comment). The comparison tests write their checks with WriteCheck, the
// validation report (go/validation/cmd/report) adds the check against NetLogo and writes the comment.
type Check struct {
	Name     string // file name and position in the comment, one of the names of Questions
	Question string
	Status   string // "✅", "❌", or empty if the check only reports (speed)
	Result   string // one line
	Scope    string // what was run and compared, one line
	Details  string // Markdown, may be empty
}

// Questions are the checks of a step, in the order of the comment.
var Questions = []Check{
	{Name: "1-changed", Question: "Did this step change the results?"},
	{Name: "2-drift", Question: "Has the model drifted away from gobees?"},
	{Name: "3-netlogo", Question: "Does it still match NetLogo?"},
	{Name: "4-speed", Question: "How fast is it?"},
}

// Question returns the check of Questions with the given name, without a result.
func Question(name string) Check {
	for _, q := range Questions {
		if q.Name == name {
			return q
		}
	}
	panic("unknown check " + name)
}

// CommentMarker starts the pull request comment, so that the workflow finds and updates it.
const CommentMarker = "<!-- check-report -->"

// WriteCheck writes the check as <dir>/<name>.json.
func WriteCheck(dir string, c Check) error {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, c.Name+".json"), b, 0o644)
}

// ReadChecks reads the checks written by WriteCheck in dir. A missing folder gives no checks.
func ReadChecks(dir string) ([]Check, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	checks := []Check{}
	for _, path := range paths {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var c Check
		if err := json.Unmarshal(b, &c); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		checks = append(checks, c)
	}
	return checks, nil
}

// Comment returns the pull request comment: one row per question of Questions, the scope of all checks and the
// details of each. A question without a check failed to give a result and is shown as failed. runURL links to the
// workflow run, may be empty.
func Comment(checks []Check, runURL string) string {
	byName := map[string]Check{}
	for _, c := range checks {
		byName[c.Name] = c
	}
	missing := "no result"
	if runURL != "" {
		missing = fmt.Sprintf("no result, see the [run](%s)", runURL)
	}
	ordered := make([]Check, len(Questions))
	for i, q := range Questions {
		c, ok := byName[q.Name]
		if !ok {
			c = q
			c.Status, c.Result = "❌", missing
		}
		ordered[i] = c
	}

	var b strings.Builder
	b.WriteString(CommentMarker + "\n## Check of this step\n\n| | Question | Result |\n| --- | --- | --- |\n")
	for _, c := range ordered {
		fmt.Fprintf(&b, "| %s | %s | %s |\n", c.Status, c.Question, c.Result)
	}
	b.WriteString("\n<details><summary>Runs</summary>\n\n")
	for _, c := range ordered {
		if c.Scope != "" {
			fmt.Fprintf(&b, "- **%s** %s\n", c.Question, c.Scope)
		}
	}
	b.WriteString("\n</details>\n")
	for _, c := range ordered {
		if c.Details != "" {
			fmt.Fprintf(&b, "\n<details><summary>%s</summary>\n\n%s\n\n</details>\n", c.Question,
				strings.TrimSpace(c.Details))
		}
	}
	if runURL != "" {
		fmt.Fprintf(&b, "\nLogs of all tests: [run](%s)\n", runURL)
	}
	return b.String()
}
