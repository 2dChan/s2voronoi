// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

package utils

import (
	"math"
	"testing"

	"github.com/golang/geo/s2"
	"github.com/google/go-cmp/cmp"
)

func TestGenerateRandomPoints_Length(t *testing.T) {
	tests := []struct {
		name string
		cnt  int
		seed int64
	}{
		{"zero points", 0, 42},
		{"one point", 1, 42},
		{"ten points", 10, 0},
		{"hundred points", 100, 99},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points := GenerateRandomPoints(tt.cnt, tt.seed)
			if len(points) != tt.cnt {
				t.Errorf("GenerateRandomPoints(%v, %v) len = %v, want %v", tt.cnt, tt.seed,
					len(points), tt.cnt)
			}
		})
	}
}

func TestGenerateRandomPoints_OnUnitSphere(t *testing.T) {
	const (
		cnt     = 100
		seed    = 0
		epsilon = 1e-12
	)
	points := GenerateRandomPoints(cnt, seed)
	for i, p := range points {
		norm := p.Norm()
		if math.Abs(norm-1.0) > epsilon {
			t.Errorf("GenerateRandomPoints(%v, %v)[%d]: point norm = %v, want ≈1", cnt, seed,
				i, norm)
		}
	}
}

func TestGenerateRandomPoints_Determinism(t *testing.T) {
	const (
		cnt  = 10
		seed = 0
	)
	a := GenerateRandomPoints(cnt, seed)
	b := GenerateRandomPoints(cnt, seed)
	if diff := cmp.Diff(b, a, cmp.AllowUnexported(s2.Point{})); diff != "" {
		t.Errorf("GenerateRandomPoints(%v, %v) mismatch (-want +got):\n%v", cnt, seed, diff)
	}
}
