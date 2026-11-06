package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
)

type cli struct {
	Convert  convertCmd  `cmd:"" help:"Convert a NetLogo parameter file and the files it names into a run file and its landscape files."`
	Generate generateCmd `cmd:"" help:"Write the run files of the repository (example farm of go/input, scenarios of go/compare)."`
}

type convertCmd struct {
	Parameters string `short:"p" type:"path" required:"" placeholder:"PATH" help:"NetLogo parameter file (MyParametersFile)."`
	Dir        string `short:"d" type:"path" default:"../../netlogo" placeholder:"PATH" help:"Folder of the input files, the folder of the .nlogo in NetLogo. The file names in the parameter file are relative to it."`
	Map        string `short:"m" default:"${default_map}" placeholder:"NAME" help:"Map (MyMap). Only the name is used: the foodsources are read from <NAME without .png>_Foodsources.txt in --dir [PAR-05]."`
	Out        string `short:"o" type:"path" required:"" placeholder:"PATH" help:"Folder of the run file and its landscape files, created if missing."`
	RunFile    string `name:"run" default:"run.yaml" help:"Name of the run file in --out."`
}

func (c *convertCmd) Run() error {
	converted, err := Convert(c.Dir, c.Parameters, c.Map)
	if err != nil {
		return err
	}
	files, err := converted.Files(c.RunFile)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(c.Out, 0o755); err != nil {
		return err
	}
	if err := writeFiles(c.Out, files); err != nil {
		return err
	}
	for _, line := range converted.Report {
		fmt.Println(line)
	}
	fmt.Printf("%d files written to %s\n", len(files), c.Out)
	return nil
}

type generateCmd struct {
	Root string `type:"path" default:"../.." placeholder:"PATH" help:"Root of the gobees-fast repository."`
}

func (c *generateCmd) Run() error {
	files, report, err := Generate()
	if err != nil {
		return err
	}
	for name, content := range files {
		path := filepath.Join(c.Root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return err
		}
	}
	for _, line := range report {
		fmt.Println(line)
	}
	fmt.Printf("%d files written\n", len(files))
	return nil
}

func main() {
	ctx := kong.Parse(&cli{}, kong.Name("netlogo2gobees"),
		kong.Description("Converts NetLogo input files of BEE-STEWARD into run files of gobees-fast."),
		kong.UsageOnError(), kong.Vars{"default_map": DefaultMap})
	ctx.FatalIfErrorf(ctx.Run())
}
