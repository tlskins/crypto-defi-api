package main

import (
	"github.com/davecgh/go-spew/spew"

	"github.com/my_projects/sol-arb-api/api"
)

func main() {
	spew.Dump(api.RoundToStr(23294.24982948924, 5))
	spew.Dump(api.RoundToStr(3232.0, 5))
	spew.Dump(api.RoundToStr(3.0, 3))
	spew.Dump(api.RoundToStr(0.0003498, 5))
	spew.Dump(api.RoundToStr(20.0003488, 6))
}
