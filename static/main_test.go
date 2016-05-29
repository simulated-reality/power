package static

import (
	"testing"

	"github.com/ready-steady/assert"
)

func TestCompute(t *testing.T) {
	Q := []float64{318.15, 328.15, 338.15, 348.15, 358.15, 368.15, 378.15, 388.15, 398.15}
	C := []float64{0.5460, 0.6304, 0.7326, 0.8550, 1.0000, 1.1711, 1.3734, 1.6067, 1.8737}
	power := New(1.0, Q, C)

	assert.EqualWithin(power.Compute(358.15), 1.088, 0.001, t)
}
