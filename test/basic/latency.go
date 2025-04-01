// SPDX-FileCopyrightText: 2023-present Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package basic

import (
	"fmt"
	"time"
)

// Latency represents latency metric capable of recording producing min, max and average latency values.
type Latency struct {
	Name  string
	Min   time.Duration
	Max   time.Duration
	Total time.Duration
	Count int
}

// NewLatency crates a new latency statistic with the specified name
func NewLatency(name string) *Latency {
	return &Latency{Name: name}
}

// Add records the specified duration
func (s *Latency) Add(duration time.Duration) *Latency {
	s.Count++
	s.Total += duration
	if s.Min == 0 || duration < s.Min {
		s.Min = duration
	}
	if s.Max == 0 || duration > s.Max {
		s.Max = duration
	}
	return s
}

// Combine combines the given latency stat into this one
func (s *Latency) Combine(other *Latency) *Latency {
	s.Count += other.Count
	s.Total += other.Total
	if s.Min == 0 || other.Min < s.Min {
		s.Min = other.Min
	}
	if s.Max == 0 || other.Max > s.Max {
		s.Max = other.Max
	}
	return s
}

func (s *Latency) String() string {
	avg := float64(s.Total.Milliseconds()) / float64(s.Count)
	return fmt.Sprintf("%-20s%14d%14.2f%14d%14d", s.Name, s.Count, avg, s.Min.Milliseconds(), s.Max.Milliseconds())
}
