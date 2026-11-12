package input

import (
	"fmt"
	"io"
	"path/filepath"
	"slices"

	"go.yaml.in/yaml/v3"
)

// siteColumn is the column of the daily output that holds the site of a run (Variation.Column).
const siteColumn = "site"

// Patterns are the site files of a variation: glob patterns (path/filepath.Match), relative to the folder of the run
// file. A path without wildcards is a pattern that matches only itself. In the run file it is one pattern or a list.
type Patterns []string

// UnmarshalYAML reads one pattern or a list of patterns.
func (p *Patterns) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode {
		var s string
		if err := n.Decode(&s); err != nil {
			return err
		}
		*p = Patterns{s}
		return nil
	}
	var list []string
	if err := n.Decode(&list); err != nil {
		return fmt.Errorf("line %d: sites is a pattern or a list of patterns", n.Line)
	}
	*p = list
	return nil
}

// Site is one site of the variation sites: a site file and the landscape files it names. A site is one landscape with
// its crop rotation, so that the two vary together.
type Site struct {
	// Name is the name of the folder of the site file, the value of the column site. Letters, digits, - and _, and
	// each site of a run file has its own.
	Name string
	// File is the site file, relative to the folder of the run file (or absolute), with / as separator.
	File string
	// Landscape and CropRotation are the landscape files of the site, relative to the folder of the run file (or
	// absolute), as the keys landscape and cropRotation of the run file.
	Landscape    string
	CropRotation []string
}

// siteFile is the content of a site file (YAML). The paths are relative to the folder of the site file.
type siteFile struct {
	// Landscape is the landscape file of the start. Needed.
	Landscape string `yaml:"landscape"`
	// CropRotation are the landscape files of the crop rotation. With fewer than 2 files, there is no crop rotation.
	CropRotation []string `yaml:"cropRotation,omitempty"`
}

// hasSites reports whether one of the variations is a list of sites.
func hasSites(vs []Variation) bool {
	return slices.ContainsFunc(vs, func(v Variation) bool { return len(v.Sites) > 0 })
}

// resolveSites returns the sites of the variations vs: the files that the patterns of the variation sites match, in
// the order of the patterns, the matches of one pattern sorted by name. glob returns the matches of a pattern, open
// opens a file, both relative to the folder of the run file. A pattern without match, a file matched twice and two
// sites of the same name are errors.
func resolveSites(vs []Variation, glob func(string) ([]string, error),
	open func(string) (io.ReadCloser, error)) ([]Site, error) {
	i := slices.IndexFunc(vs, func(v Variation) bool { return len(v.Sites) > 0 })
	if i < 0 {
		return nil, nil
	}
	var sites []Site
	files := map[string]bool{}
	names := map[string]string{} // the site file by the name of its site
	for _, pattern := range vs[i].Sites {
		matches, err := glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("%s: sites: %s: %w", variationsKey, pattern, err)
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("%s: sites: %s matches no file", variationsKey, pattern)
		}
		slices.Sort(matches)
		for _, m := range matches {
			m = filepath.ToSlash(m)
			if files[m] {
				return nil, fmt.Errorf("%s: sites: %s is matched twice", variationsKey, m)
			}
			files[m] = true
			s, err := readSite(m, open)
			if err != nil {
				return nil, err
			}
			if other, ok := names[s.Name]; ok {
				return nil, fmt.Errorf("%s: sites: %s and %s have the same name %q (the name of the folder), "+
					"the column %s could not tell them apart", variationsKey, other, m, s.Name, siteColumn)
			}
			names[s.Name] = m
			sites = append(sites, s)
		}
	}
	return sites, nil
}

// readSite reads the site file name (relative to the folder of the run file). Its paths become relative to the
// folder of the run file.
func readSite(name string, open func(string) (io.ReadCloser, error)) (Site, error) {
	var sf siteFile
	if _, err := readFile(name, open, func(r io.Reader) error { return decodeStrict(r, &sf) }); err != nil {
		return Site{}, fmt.Errorf("site file %w", err)
	}
	if sf.Landscape == "" {
		return Site{}, fmt.Errorf("site file %s: landscape is missing", name)
	}
	dir := filepath.Dir(name)
	s := Site{Name: filepath.Base(dir), File: name, Landscape: sitePath(dir, sf.Landscape)}
	if err := CheckName(s.Name); err != nil {
		return Site{}, fmt.Errorf("site file %s: the folder is the name of the site: %w", name, err)
	}
	for _, l := range sf.CropRotation {
		s.CropRotation = append(s.CropRotation, sitePath(dir, l))
	}
	return s, nil
}

// sitePath returns the path name of a site file in the folder dir relative to the folder of the run file, with / as
// separator. An absolute path stays.
func sitePath(dir, name string) string {
	if filepath.IsAbs(name) {
		return filepath.ToSlash(name)
	}
	return filepath.ToSlash(filepath.Join(dir, name))
}

// setSite sets the keys landscape and cropRotation of the run file root (a mapping) to those of the site s.
func setSite(root *yaml.Node, s Site) error {
	var n yaml.Node
	if err := n.Encode(s.Landscape); err != nil {
		return err
	}
	if err := setPath(root, "landscape", &n); err != nil {
		return err
	}
	if len(s.CropRotation) == 0 {
		return nil
	}
	n = yaml.Node{}
	if err := n.Encode(s.CropRotation); err != nil {
		return err
	}
	return setPath(root, "cropRotation", &n)
}
