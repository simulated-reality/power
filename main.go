// Package power provides algorithms for distributing power consumption of
// applications running on multiprocessor platforms.
package power

import (
	"github.com/ready-steady/sort"
	"github.com/turing-complete/system"
	"github.com/turing-complete/time"
)

// Power represents a power simulator configured for a particular system.
type Power struct {
	platform    *system.Platform
	application *system.Application
}

// New returns a power distributor for a platform and an application.
func New(platform *system.Platform, application *system.Application) *Power {
	return &Power{platform: platform, application: application}
}

// Partition does what the standalone Partition function does.
func (p *Power) Partition(schedule *time.Schedule, points []float64,
	ε float64) ([]float64, []float64, []uint) {

	return Partition(p.collect(schedule), schedule, points, ε)
}

// Sample does what the standalone Sample function does.
func (p *Power) Sample(schedule *time.Schedule, Δt float64, ns uint) []float64 {
	return Sample(p.collect(schedule), schedule, Δt, ns)
}

// Progress returns a function func(time float64, power []float64) that computes
// the power consumption at an arbitrary time moment according to a schedule.
func (p *Power) Progress(schedule *time.Schedule) func(float64, []float64) {
	cores, tasks := p.platform.Cores, p.application.Tasks
	nc, nt := uint(len(cores)), uint(len(tasks))

	mapping := make([][]uint, nc)
	for i := uint(0); i < nc; i++ {
		mapping[i] = make([]uint, 0, nt)
		for j := uint(0); j < nt; j++ {
			if i == schedule.Mapping[j] {
				mapping[i] = append(mapping[i], j)
			}
		}
	}

	start, finish := schedule.Start, schedule.Finish

	return func(time float64, power []float64) {
		for i := uint(0); i < nc; i++ {
			power[i] = 0
			for _, j := range mapping[i] {
				if start[j] <= time && time <= finish[j] {
					power[i] = cores[i].Power[tasks[j].Type]
					break
				}
			}
		}
	}
}

func (p *Power) collect(schedule *time.Schedule) []float64 {
	cores, tasks := p.platform.Cores, p.application.Tasks
	nt := uint(len(tasks))

	power := make([]float64, nt)
	for i := uint(0); i < nt; i++ {
		power[i] = cores[schedule.Mapping[i]].Power[tasks[i].Type]
	}

	return power
}

// Partition computes a power profile with a variable time step dictated by the
// time moments of power switches (the start and finish times of the tasks) and
// a number of additional time moments gathered in points.
func Partition(power []float64, schedule *time.Schedule, points []float64,
	ε float64) ([]float64, []float64, []uint) {

	nc, nt, np := schedule.Cores, schedule.Tasks, uint(len(points))

	time := make([]float64, 2*nt+np)
	copy(time[:nt], schedule.Start)
	copy(time[nt:], schedule.Finish)
	copy(time[2*nt:], points)

	ΔT, steps := traverse(time, ε)
	ssteps, fsteps, psteps := steps[:nt], steps[nt:2*nt], steps[2*nt:]

	ns := uint(len(ΔT))

	P := make([]float64, nc*ns)

	for i := uint(0); i < nt; i++ {
		j := schedule.Mapping[i]
		p := power[i]

		s, f := ssteps[i], fsteps[i]

		for ; s < f; s++ {
			P[s*nc+j] = p
		}
	}

	return P, ΔT, psteps
}

// Sample computes a power profile with respect to a sampling interval Δt. The
// required number of samples is specified by ns; short schedules are extended
// while long ones are truncated.
func Sample(power []float64, schedule *time.Schedule, Δt float64, ns uint) []float64 {
	nc, nt := schedule.Cores, schedule.Tasks

	P := make([]float64, nc*ns)

	if count := uint(schedule.Span / Δt); count < ns {
		ns = count
	}

	for i := uint(0); i < nt; i++ {
		j := schedule.Mapping[i]
		p := power[i]

		s := uint(schedule.Start[i]/Δt + 0.5)
		f := uint(schedule.Finish[i]/Δt + 0.5)
		if f > ns {
			f = ns
		}

		for ; s < f; s++ {
			P[s*nc+j] = p
		}
	}

	return P
}

func traverse(points []float64, ε float64) ([]float64, []uint) {
	np := uint(len(points))
	order, _ := sort.Quick(points)

	Δ := make([]float64, np-1)
	steps := make([]uint, np)

	j := uint(0)

	for i, x := uint(1), points[0]; i < np; i++ {
		if δ := points[i] - x; δ > ε {
			x = points[i]
			Δ[j] = δ
			j++
		}
		steps[order[i]] = j
	}

	return Δ[:j], steps
}
