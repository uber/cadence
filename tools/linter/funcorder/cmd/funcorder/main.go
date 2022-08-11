package main

import (
	"github.com/uber/cadence/tools/linter/funcorder"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(funcorder.Analyzer)
}
