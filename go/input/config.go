package input

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"

	"gobees/params"

	"go.yaml.in/yaml/v3"
)

// Config is the content of a run file (YAML). The tool tools/netlogo2gobees converts the NetLogo input files
// into run files.
//
// All keys are lowerCamelCase, the NetLogo names are in docs/names.md and in the tag `netlogo` of the fields that
// the converter maps by name. A key missing in the file keeps its zero value, except in the section constants,
// which must hold every key. An unknown key is an error.
type Config struct {
	// Run holds the seed and the days of the run and, in a run info, the record of a run. The section is optional.
	Run *Run `yaml:"run,omitempty"`

	Parameters params.Parameters `yaml:"parameters"`

	// Output are the options of the output. NetLogo: the parameter KeepDeadColonies?.
	Output params.Output `yaml:"output"`

	// Constants are the values that NetLogo hard-codes and its parameter files cannot change. The section must hold
	// every key (checkComplete), so that no value of a run comes from the code unseen.
	Constants params.Constants `yaml:"constants"`

	// InitialQueens are the species and their initial queens, in this order (NetLogo:
	// BeeSpeciesInitialQueensListAsString).
	InitialQueens []params.SpeciesQueens `yaml:"initialQueens"`

	// ForagingPeriod is the foraging period per day of the year [s], 365 values per year, for one or more years.
	// The years repeat (NetLogo: Weather).
	ForagingPeriod []Year `yaml:"foragingPeriod"`

	// StewardshipOptions are the names of the stewardship options (NetLogo: the habitat types of the HabitatsFile
	// whose habitatColourID contains "FlowerSpeciesList_").
	StewardshipOptions []string `yaml:"stewardshipOptions"`

	// Species are the bee species (NetLogo: rows of the BeespeciesFile). Only the species of InitialQueens are
	// used, in the order of this list, found by their name.
	Species []Species `yaml:"species"`

	// Flowers are the flower species (NetLogo: rows of the FlowerspeciesFile). Without the key, no layers
	// get flower values.
	Flowers []params.FlowerSpecies `yaml:"flowers"`

	// Landscape is the landscape file of the start, relative to the folder of the run file (NetLogo: the
	// foodsources file of MyMap).
	Landscape string `yaml:"landscape"`

	// CropRotation are the landscape files of the crop rotation, relative to the folder of the run file.
	// With fewer than 2 files, there is no crop rotation (NetLogo: CropRotationListAsString).
	CropRotation []string `yaml:"cropRotation,omitempty"`
}

// Year holds the foraging period of the 365 days of one year [s]. It is written as one line.
type Year []float64

// MarshalYAML writes the year as a flow sequence, i.e. on one line.
func (y Year) MarshalYAML() (any, error) {
	return flowNode([]float64(y))
}

// Species are the values of one bee species in a run file: params.SpeciesFile, which params.Species embeds.
type Species = params.SpeciesFile

// Landscape is the content of a landscape file (YAML) (NetLogo: foodsources file).
type Landscape struct {
	MetresPerPatch float64 `yaml:"metresPerPatch"` // metres per NetLogo patch
	Patches        []Patch `yaml:"patches"`        // flower patches, in this order
}

// Patch is one flower patch of a landscape file. The keys are the columns of the NetLogo foodsources file, except
// area (size_sqm) and flowers (flowerSpeciesList). The keys x and y would need quotes in YAML (y is a boolean in
// YAML 1.1), so xcor and ycor stay.
type Patch struct {
	ID        float64    `yaml:"id"`
	PatchType string     `yaml:"patchType"`
	X         float64    `yaml:"xcor"`
	Y         float64    `yaml:"ycor"`
	Area      float64    `yaml:"area"` // [m2]
	Flowers   FlowerList `yaml:"flowers"`
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
	if err := checkWhole(&root, reflect.TypeOf(v).Elem(), ""); err != nil {
		return err
	}
	if _, ok := v.(*Config); ok {
		return checkConstants(&root)
	}
	return nil
}

// checkConstants checks that the run file root holds the section constants with every key of params.Constants.
func checkConstants(root *yaml.Node) error {
	n := mappingValue(root.Content[0], constantsKey)
	if n == nil {
		return fmt.Errorf("%s: missing, the section must hold every constant (the converter writes it)", constantsKey)
	}
	return checkComplete(n, reflect.TypeFor[params.Constants](), constantsKey)
}

// checkComplete checks that the mapping n holds every key of the struct type t, recursively. path is the key of
// n in the error messages.
func checkComplete(n *yaml.Node, t reflect.Type, path string) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: %s: expected a mapping", n.Line, path)
	}
	for i := range t.NumField() {
		f := t.Field(i)
		key, _, _ := strings.Cut(f.Tag.Get("yaml"), ",")
		value := mappingValue(n, key)
		if value == nil {
			return fmt.Errorf("line %d: %s.%s: missing, the section %s must hold every key", n.Line, path, key, constantsKey)
		}
		if f.Type.Kind() == reflect.Struct {
			if err := checkComplete(value, f.Type, path+"."+key); err != nil {
				return err
			}
		}
	}
	return nil
}

// mappingValue returns the value of key in the mapping n, or nil.
func mappingValue(n *yaml.Node, key string) *yaml.Node {
	if n.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if n.Content[i].Value == key {
			return n.Content[i+1]
		}
	}
	return nil
}

// constantsKey is the key of the section constants in a run file.
const constantsKey = "constants"

// runKey is the key of the section run in a run file.
const runKey = "run"

// constantsComment is the comment above the section constants.
const constantsComment = "Fixed in the NetLogo code, not settable in NetLogo's parameter files. Defaults: the NetLogo values.\n" +
	"Every key is needed. The comment at a key gives unit and source."

// WriteConfig writes a run file. The section constants gets a comment above it and, at each key, the tag doc of the
// field (unit and source). The keys seed, days and runs of the section run get the tag doc of their field.
func WriteConfig(w io.Writer, c Config) error {
	return WriteConfigNotes(w, c, nil)
}

// WriteConfigNotes writes a run file as WriteConfig and adds notes to the comments of keys of the section
// constants. notes holds the note by the path of the key below constants, e.g. "foraging.maxForagingRange".
func WriteConfigNotes(w io.Writer, c Config, notes map[string]string) error {
	n := &yaml.Node{}
	if err := n.Encode(c); err != nil {
		return err
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		switch n.Content[i].Value {
		case constantsKey:
			n.Content[i].HeadComment = constantsComment
			commentKeys(n.Content[i+1], reflect.TypeFor[params.Constants](), "", notes)
		case runKey:
			commentKeys(n.Content[i+1], reflect.TypeFor[Run](), "", nil)
		}
	}
	return encode(w, n)
}

// commentKeys sets the tag doc of each field of the struct type t as line comment of its key in the mapping n,
// recursively, with the note of notes for the path of the key appended.
func commentKeys(n *yaml.Node, t reflect.Type, path string, notes map[string]string) {
	for i := range t.NumField() {
		f := t.Field(i)
		key, _, _ := strings.Cut(f.Tag.Get("yaml"), ",")
		keyPath := key
		if path != "" {
			keyPath = path + "." + key
		}
		value := mappingValue(n, key)
		if value == nil {
			continue
		}
		if f.Type.Kind() == reflect.Struct {
			commentKeys(value, f.Type, keyPath, notes)
			continue
		}
		comment := f.Tag.Get("doc")
		if note := notes[keyPath]; note != "" {
			comment = strings.TrimPrefix(comment+". "+note, ". ")
		}
		value.LineComment = comment
	}
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
