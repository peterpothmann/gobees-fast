package input

import "fmt"

// Run is the section run of a run file: the seed, the days, the number of runs and of threads, and a record of how a
// run went. The section and its keys seed, runs and threads are needed, each 1 or more (checkRun): no value of a run
// is drawn or taken from the machine unseen. The converter writes seed 1, 365 days, 1 run and 1 thread. gobees writes
// the section into the run info next to the daily output (main.go), so that the run info repeats the run as a run
// file. Build reads Seed and Days, main.go reads Runs and Threads, the other keys are a record. The columns of the
// output are in the key columns of the section output.
type Run struct {
	// Seed is the random seed of the first run, 1 or more. The run i (from 0) has the seed Seed + i. The flag -s
	// overrides it.
	Seed int64 `yaml:"seed" doc:"1 or more. Run i (from 1) has the seed seed + i - 1"`
	// Days is the number of days (ticks) to run. 0: no limit. No key: params.Default. The flag --days overrides it.
	Days *int `yaml:"days,omitempty" doc:"[d] 0: no limit, the run ends when the population is extinct (stopExtinct)"`
	// Runs is the number of runs, 1 or more. The flag --runs overrides it.
	Runs int `yaml:"runs" doc:"1 or more, the daily output holds all runs (columns run and seed)"`
	// Threads is the number of runs at the same time, 1 or more. It does not change the output. On a cluster: the CPUs
	// of the job. The flag --threads overrides it.
	Threads int `yaml:"threads" doc:"runs at the same time, 1 or more. Does not change the output"`

	// Started is the start time of the run (RFC 3339).
	Started string `yaml:"started,omitempty"`
	// Command is the command line of the run.
	Command []string `yaml:"command,omitempty,flow"`
	// Version is the build of gobees that ran.
	Version *Version `yaml:"version,omitempty"`
	// Config is the run file given with -c, or "built-in example farm".
	Config string `yaml:"config,omitempty"`
	// Output is the daily output file.
	Output string `yaml:"output,omitempty"`
	// Landscapes are the copies of the landscape files next to the run info.
	Landscapes []LandscapeCopy `yaml:"landscapes,omitempty"`
}

// Version is the build of gobees, from runtime/debug.ReadBuildInfo. The keys of the version control system are
// missing without that information, e.g. with go run.
type Version struct {
	Module   string `yaml:"module"`             // module path
	Version  string `yaml:"version,omitempty"`  // module version, e.g. (devel)
	Revision string `yaml:"revision,omitempty"` // VCS revision (commit)
	Time     string `yaml:"time,omitempty"`     // time of the revision
	Modified string `yaml:"modified,omitempty"` // "true" if the working tree had changes against the revision
	Go       string `yaml:"go"`                 // Go version of the build
}

// LandscapeCopy is a landscape file of a run, copied byte for byte next to the run info.
type LandscapeCopy struct {
	File   string `yaml:"file"`   // name of the copy, in the keys landscape and cropRotation of the run info
	Source string `yaml:"source"` // name of the landscape file in the run file of the run
	SHA256 string `yaml:"sha256"` // SHA-256 of the file, hexadecimal
}

// checkRun checks the section run of a run file: it must be there, with seed, runs and threads of 1 or more. A missing
// key reads as 0 and fails the same way.
func checkRun(r *Run) error {
	if r == nil {
		return fmt.Errorf("section run is missing: it needs seed, runs and threads, each 1 or more")
	}
	for _, k := range []struct {
		key   string
		value int64
	}{{"seed", r.Seed}, {"runs", int64(r.Runs)}, {"threads", int64(r.Threads)}} {
		if k.value < 1 {
			return fmt.Errorf("run.%s is %d or missing, want 1 or more", k.key, k.value)
		}
	}
	if r.Days != nil && *r.Days < 0 {
		return fmt.Errorf("run.days is %d, want 0 (no limit) or more", *r.Days)
	}
	return nil
}
