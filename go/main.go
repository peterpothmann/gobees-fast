package main

import (
	"fmt"

	"gobees/globals"
	"gobees/model"
	"gobees/params"

	"github.com/mlange-42/ark/ecs"
)

func main() {
	p := params.Default()
	a := model.Default(p, nil)
	a.Run()

	ticks := ecs.GetResource[globals.Ticks](a.World)
	fmt.Printf("finished after %d days\n", ticks.Ticks)
}
