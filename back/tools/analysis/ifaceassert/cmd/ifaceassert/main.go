package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"klubok/tools/analysis/ifaceassert"
)

func main() {
	singlechecker.Main(ifaceassert.New())
}
