package input

// Run is the section run of a run file: the seed and the days of the run, and a record of how a run went. gobees
// writes it into the run info next to the daily output (main.go), so that the run info repeats the run as a run
// file. Build reads only Seed and Days, the other keys are a record.
type Run struct {
	// Seed is the random seed. 0 or no key: a random seed. The flag -s overrides it.
	Seed int64 `yaml:"seed,omitempty"`
	// Days is the number of days (ticks) to run. 0: no limit. No key: params.Default. The flag --days overrides it.
	Days *int `yaml:"days,omitempty"`

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
	// Columns are the columns of the daily output, in this order.
	Columns []string `yaml:"columns,omitempty,flow"`
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
