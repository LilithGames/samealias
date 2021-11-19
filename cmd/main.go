package main

import (
	"github.com/LilithGames/samealias"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(samealias.Analyzer)
}
