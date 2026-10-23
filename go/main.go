package main

import (
	"fmt"
	"log"
	"path/filepath"

	"gobees/globals"
	"gobees/model"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

func main() {
	p := params.Default()
	p.WorkingDirectory.Path = filepath.Join("..", "netlogo")

	// [INIT-04] With the default interface settings, NetLogo reads MyParametersFile (see [INIT-02]).
	file, err := params.ReadParameterFile(filepath.Join(p.WorkingDirectory.Path, "_SYSTEM_Example_Farm_Parameters.csv"))
	if err != nil {
		log.Fatal(err)
	}
	p.File = file

	a := model.Default(p, nil)
	a.Run()

	ticks := ecs.GetResource[globals.Ticks](a.World)
	derived := ecs.GetResource[globals.Derived](a.World)
	counters := ecs.GetResource[globals.Counters](a.World)
	fmt.Printf("species: %v\n", derived.BeeSpeciesInitialQueensList)
	fmt.Printf("colonies founded: %v, bees died: %v\n", counters.TotalColoniesEverProduced, counters.TotalBeesEverDied)
	fmt.Printf("finished after %d days\n", ticks.Ticks)
}
