// Package leakage provides algorithms for calculating the leakage power.
package leakage

import (
	"github.com/ready-steady/statistics/regression"
)

// Power is a power calculator.
type Power struct {
	nominal float64
	model   *regression.SimpleLinear
}

// New returns a power calculator.
func New(nominal float64, temperature, coefficient []float64) *Power {
	return &Power{
		nominal: nominal,
		model:   regression.NewSimpleLinear(temperature, coefficient),
	}
}

func (self *Power) Compute(temperature float64) float64 {
	return self.nominal * self.model.Compute(temperature)
}
