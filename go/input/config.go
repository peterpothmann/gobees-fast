package input

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"

	"gobees/day"
	"gobees/globals"
	"gobees/params"

	"go.yaml.in/yaml/v3"
)

// Config is the content of a run file (YAML) [INIT-04]. The tool tools/netlogo2gobees converts the NetLogo input files
// into run files.
//
// Keys of NetLogo values keep their NetLogo names (parameter and column names). The other keys are new. A key
// missing in the file keeps its zero value (as a column missing in a NetLogo parameter file), an unknown key is an
// error.
type Config struct {
	Parameters params.ParameterFile `yaml:"parameters"`

	// Output are the options of the output (step F13). NetLogo: the parameter KeepDeadColonies?.
	Output params.Output `yaml:"output"`

	// InitialQueens are the species and their initial queens, in this order [QINI] (NetLogo:
	// BeeSpeciesInitialQueensListAsString).
	InitialQueens []globals.SpeciesQueens `yaml:"initial_queens"`

	// ForagingPeriod_s is the foraging period per day of the year [s], 365 values per year, for one or more years.
	// The years repeat [FPER] (NetLogo: Weather).
	ForagingPeriod_s []Year `yaml:"foraging_period_s"`

	// StewardshipOptions are the names of the stewardship options [CSO] (NetLogo: the habitat types of the HabitatsFile
	// whose habitatColourID contains "FlowerSpeciesList_").
	StewardshipOptions []string `yaml:"stewardship_options"`

	// Species are the bee species [SPEC] (NetLogo: rows of the BeespeciesFile). Only the species of InitialQueens are
	// used, in the order of this list, found by their name [SPEC-02].
	Species []Species `yaml:"species"`

	// Flowers are the flower species [LAYR-02] (NetLogo: rows of the FlowerspeciesFile). Without the key, no layers
	// get flower values (NetLogo: FlowerspeciesFile "No Input File" [FSRC-12]).
	Flowers []params.FlowerSpecies `yaml:"flowers"`

	// Landscape is the landscape file of the start [PAR-05], relative to the folder of the run file (NetLogo: the
	// foodsources file of MyMap).
	Landscape string `yaml:"landscape"`

	// CropRotation are the landscape files of the crop rotation [PAR-07], relative to the folder of the run file.
	// With fewer than 2 files, there is no crop rotation [MORN-09] (NetLogo: CropRotationListAsString).
	CropRotation []string `yaml:"crop_rotation,omitempty"`
}

// Year holds the foraging period of the 365 days of one year [s]. It is written as one line.
type Year []float64

// MarshalYAML writes the year as a flow sequence, i.e. on one line.
func (y Year) MarshalYAML() (any, error) {
	return flowNode([]float64(y))
}

// Species are the values of one bee species in a run file [SPEC-03], before the derived values. The keys are the
// column names of the NetLogo BeespeciesFile.
type Species struct {
	// Name is the name in InitialQueens [SPEC-02], [IQUE-04]. NetLogo has two columns, species (compared with the
	// initial queens) and name (stored), equal in all files. The converter checks that they are equal (step F13).
	Name string `yaml:"name"`
	// Parasitic marks a cuckoo bumblebee species (Psithyrus) [ACT-04], [EMER-04], [PSI]. NetLogo compares the name
	// with "Psithyrus", the converter sets the key for this name (step F13).
	Parasitic                    bool       `yaml:"parasitic,omitempty"`
	EmergingDay_mean             float64    `yaml:"emergingDay_mean"`
	EmergingDay_sd               float64    `yaml:"emergingDay_sd"`
	NestHabitatsList             []string   `yaml:"nestHabitatsList"` // [SPEC-04] before the expansion
	Proboscis_min_mm             float64    `yaml:"proboscis_min_mm"`
	Proboscis_max_mm             float64    `yaml:"proboscis_max_mm"`
	GrowthFactor                 float64    `yaml:"growthFactor"`
	SeasonStop                   day.OfYear `yaml:"seasonStop"`
	MaxLifespanWorkers           int        `yaml:"maxLifespanWorkers"`
	Batchsize                    int        `yaml:"batchsize"`
	FlightVelocity_m_per_s       float64    `yaml:"flightVelocity_m/s"`
	SearchLength_m               float64    `yaml:"searchLength_m"`
	TimeUnloading                float64    `yaml:"timeUnloading"`
	SpecMax_cropVolume_myl       float64    `yaml:"specMax_cropVolume_myl"`
	SpecMax_pollenPellets_g      float64    `yaml:"specMax_pollenPellets_g"`
	MinToMaxFactor               float64    `yaml:"minToMaxFactor"`
	DevAgeHatchingMin_d          float64    `yaml:"devAgeHatchingMin_d"`
	DevAgePupationMin_d          float64    `yaml:"devAgePupationMin_d"`
	DevAgeEmergingMin_d          float64    `yaml:"devAgeEmergingMin_d"`
	DevWeightEgg_mg              float64    `yaml:"devWeightEgg_mg"`
	DevWeightPupationMin_mg      float64    `yaml:"devWeightPupationMin_mg"`
	DevWeightPupationMax_mg      float64    `yaml:"devWeightPupationMax_mg"`
	PollenToBodymassFactor       float64    `yaml:"pollenToBodymassFactor"`
	Dev_Q_DeterminationWeight_mg float64    `yaml:"dev_Q_DeterminationWeight_mg"`
	DevAge_Q_PupationMin_d       float64    `yaml:"devAge_Q_PupationMin_d"`
	DevWeight_Q_PupationMin_mg   float64    `yaml:"devWeight_Q_PupationMin_mg"`
	DevWeight_Q_PupationMax_mg   float64    `yaml:"devWeight_Q_PupationMax_mg"`
	DevAge_Q_EmergingMin_d       float64    `yaml:"devAge_Q_EmergingMin_d"`
	DailyNestSiteChance          float64    `yaml:"dailyNestSiteChance"`
}

// Landscape is the content of a landscape file (YAML) [FSRC-02] to [FSRC-04] (NetLogo: foodsources file).
type Landscape struct {
	MetresPerPatch float64 `yaml:"metres_per_patch"` // metres per NetLogo patch [FSRC-02] (NetLogo: line 1)
	Patches        []Patch `yaml:"patches"`          // flower patches, in this order [FSRC-04]
}

// Patch is one flower patch of a landscape file. The keys are the column names of the NetLogo foodsources file.
type Patch struct {
	Id                float64    `yaml:"id"`
	PatchType         string     `yaml:"patchType"`
	X                 float64    `yaml:"xcor"`
	Y                 float64    `yaml:"ycor"`
	Area_sqm          float64    `yaml:"size_sqm"`
	FlowerSpeciesList FlowerList `yaml:"flowerSpeciesList"`
}

// FlowerList is the flowerSpeciesList of a flower patch. It is written on one line, as in NetLogo.
type FlowerList []FlowerEntry

// MarshalYAML writes the list as a flow sequence.
func (l FlowerList) MarshalYAML() (any, error) {
	return flowNode([]FlowerEntry(l))
}

// FlowerEntry is one item of a flowerSpeciesList, written as in NetLogo: [Bramble, 0.331] or [Margin].
type FlowerEntry params.FlowerEntry

// MarshalYAML writes the entry as a flow sequence of its name and, if it has one, its value.
func (e FlowerEntry) MarshalYAML() (any, error) {
	if e.HasValue {
		return flowNode([]any{e.Name, e.Value})
	}
	return flowNode([]any{e.Name})
}

// UnmarshalYAML reads [name] or [name, value].
func (e *FlowerEntry) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.SequenceNode || len(n.Content) < 1 || len(n.Content) > 2 {
		return fmt.Errorf("line %d: a flower entry is [name] or [name, value]", n.Line)
	}
	*e = FlowerEntry{}
	if err := n.Content[0].Decode(&e.Name); err != nil {
		return err
	}
	if len(n.Content) == 2 {
		e.HasValue = true
		return n.Content[1].Decode(&e.Value)
	}
	return nil
}

// flowNode returns v as a YAML node in flow style, i.e. on one line.
func flowNode(v any) (*yaml.Node, error) {
	n := &yaml.Node{}
	if err := n.Encode(v); err != nil {
		return nil, err
	}
	n.Style = yaml.FlowStyle
	return n, nil
}

// ReadConfig reads a run file. An unknown key or a second document is an error.
func ReadConfig(r io.Reader) (Config, error) {
	c := Config{}
	return c, decodeStrict(r, &c)
}

// ReadLandscape reads a landscape file. An unknown key or a second document is an error.
func ReadLandscape(r io.Reader) (Landscape, error) {
	l := Landscape{}
	return l, decodeStrict(r, &l)
}

// decodeStrict decodes one YAML document into v (a pointer). An unknown key, a second document and a number with a
// fraction for an integer field (checkWhole) are errors.
func decodeStrict(r io.Reader, v any) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	d := yaml.NewDecoder(bytes.NewReader(data))
	d.KnownFields(true)
	if err := d.Decode(v); err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("empty file")
		}
		return err
	}
	var extra any
	if err := d.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("more than one YAML document")
	}
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return err
	}
	return checkWhole(&root, reflect.TypeOf(v).Elem(), "")
}

// WriteConfig writes a run file.
func WriteConfig(w io.Writer, c Config) error {
	return encode(w, c)
}

// WriteLandscape writes a landscape file.
func WriteLandscape(w io.Writer, l Landscape) error {
	return encode(w, l)
}

func encode(w io.Writer, v any) error {
	var b bytes.Buffer
	e := yaml.NewEncoder(&b)
	e.SetIndent(2)
	if err := e.Encode(v); err != nil {
		return err
	}
	if err := e.Close(); err != nil {
		return err
	}
	_, err := w.Write(b.Bytes())
	return err
}
