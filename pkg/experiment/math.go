package experiment

import "math"

func average(x []float64) float64 {
	var avr float64
	for _, val := range x {
		avr += val
	}
	avr /= float64(len(x))
	return avr
}
func variance(x []float64) float64 {

	avr := average(x)
	variance := float64(0)
	for _, val := range x {
		variance += math.Sqrt(math.Abs(avr - val))
	}
	variance /= float64(len(x))
	return variance
}
