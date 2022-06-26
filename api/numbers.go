package api

import (
	"fmt"
	"math"
	"strconv"
)

func RoundToStr(val float64, precision int) string {
	wholesLen := len(strconv.Itoa(int(val)))
	roundStr := strconv.Itoa(int(math.Round(val * math.Pow10(precision))))
	for len(roundStr) <= precision {
		roundStr = "0" + roundStr
	}
	wholes := roundStr[:wholesLen]
	decs := roundStr[wholesLen:]
	for len(decs) < precision {
		decs = decs + "0"
	}
	return fmt.Sprintf("%s.%s", wholes, decs)
}
