package distribution

import (
	"fmt"
	"math"
)

type Distribution struct {
	N     float64
	Min   float64
	Max   float64
	SumX  float64
	SumX2 float64
}

func FromValue(v float64) *Distribution {
	return &Distribution{1, v, v, v, v * v}
}

func (d *Distribution) AddEntry(v float64) {
	d.Min = math.Min(d.Min, v)
	d.Max = math.Max(d.Max, v)
	d.SumX += v
	d.SumX2 += v * v
	d.N += 1
}

func (d *Distribution) Add(dist *Distribution) {
	d.Min = math.Min(d.Min, dist.Min)
	d.Max = math.Max(d.Max, dist.Max)
	d.SumX += dist.SumX
	d.SumX2 += dist.SumX2
	d.N += dist.N
}

func (d *Distribution) Mean() float64 {
	return d.SumX / float64(d.N)
}

func (d *Distribution) Sd() float64 {
	meanXSq := d.SumX2 / d.N
	meanSq := math.Pow(d.Mean(), 2)
	return math.Sqrt(math.Max(meanXSq-meanSq, 0))
}

func (d *Distribution) String() string {
	return fmt.Sprintf("Distribution: mean: %.4g, sd: %.4g, max: %.4g, min: %.4g (weight %.4g)", d.Mean(), d.Sd(), d.Max, d.Min, d.N)
}
