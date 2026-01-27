// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

package s2voronoi

import (
	"fmt"
	"math"
	"testing"

	"github.com/2dChan/s2voronoi/utils"
	"github.com/golang/geo/s2"
)

// DiagramOptions

func TestWithEps(t *testing.T) {
	tests := []struct {
		name    string
		eps     float64
		wantErr bool
	}{
		{"eps positive", 0.5, false},
		{"eps zero", 0, true},
		{"eps negative", -1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &DiagramOptions{Eps: defaultEps}
			opt := WithEps(tt.eps)
			err := opt(opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("WithEps(%v) error = %v, wantErr %v", tt.eps, err, tt.wantErr)
			}
			if err == nil && opts.Eps != tt.eps {
				t.Errorf("WithEps(%v) opts.Eps = %v, want %v", tt.eps, opts.Eps, tt.eps)
			}
		})
	}
}

// Diagram

func TestNewDiagram_WithEps(t *testing.T) {
	points := utils.GenerateRandomPoints(10, 0)
	tests := []struct {
		name    string
		eps     float64
		wantErr bool
	}{
		{"eps positive small", 0.01, false},
		{"eps zero", 0, true},
		{"eps negative", -0.01, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDiagram(points, WithEps(tt.eps))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDiagram(..., WithEps(%v)) error = %v, wantErr %v", tt.eps, err, tt.wantErr)
			}
		})
	}
}

func TestDiagram_Invariants(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"minimal", 4},
		{"small", 10},
		{"medium", 1000},
		{"large", 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vd := mustNewDiagram(t, tt.size)

			// Euler's formula for spherical Voronoi Diagram: V = 2n - 4
			want := 2*tt.size - 4
			got := len(vd.Vertices)
			if got != want {
				t.Errorf("vd.Vertices count = %v, want %v", got, want)
			}

			want1 := tt.size
			got1 := len(vd.Sites)
			if got1 != want1 {
				t.Errorf("vd.Sites count = %v, want %v", got1, want1)
			}

			want2 := len(vd.Sites)
			got2 := vd.NumCells()
			if got2 != want2 {
				t.Errorf("vd.NumCells() = %v, want %v", got2, want2)
			}
		})
	}
}

func TestNewTriangulation_DegenerateInput(t *testing.T) {
	// TODO: Add more tests for broken or invalid scenarios.
	points := utils.GenerateRandomPoints(3, 0)
	if _, err := NewDiagram(points); err == nil {
		t.Errorf("NewDiagram(...) error = nil, want non-nil")
	}
}

func TestNewDiagram_OnSphere(t *testing.T) {
	vd := mustNewDiagram(t, 100)

	for i, v := range vd.Vertices {
		n := v.Norm()
		if math.Abs(n-1.0) > defaultEps {
			t.Errorf("vd.Vertices[%d] norm = %v, want ~1.0", i, n)
		}
	}

	for i, s := range vd.Sites {
		n := s.Norm()
		if math.Abs(n-1.0) > defaultEps {
			t.Errorf("vd.Sites[%d] norm = %v, want ~1.0", i, n)
		}
	}
}

func TestNewDiagram_VerifyCCW(t *testing.T) {
	vd := mustNewDiagram(t, 100)

	for i := range vd.NumCells() {
		cell, err := vd.Cell(i)
		if err != nil {
			t.Fatalf("vd.Cell(%d) error = %v, want nil", i, err)
		}

		center := cell.Site()
		for i := 0; i < cell.NumVertices(); i++ {
			cIdx := i
			nIdx := (i + 1) % cell.NumVertices()
			c, err := cell.Vertex(cIdx)
			if err != nil {
				t.Fatalf("cell.Vertex(%d) error = %v, want nil", cIdx, err)
			}
			n, err := cell.Vertex(nIdx)
			if err != nil {
				t.Fatalf("cell.Vertex(%d) error = %v, want nil", nIdx, err)
			}

			angle := computeAngleCCW(c, n, center)
			if angle <= 0 {
				t.Errorf("vd.Cell(%d) Vertices %d,%d not sorted in CCW", i,
					cIdx, nIdx)
			}
		}

		for i := 0; i < cell.NumNeighbors(); i++ {
			cIdx := i
			nIdx := (i + 1) % cell.NumNeighbors()
			neigh, err := cell.Neighbor(cIdx)
			if err != nil {
				t.Fatalf("cell.Neighbor(%d) error = %v, want nil", cIdx, err)
			}
			c := neigh.Site()
			neigh2, err := cell.Neighbor(nIdx)
			if err != nil {
				t.Fatalf("cell.Neighbor(%d) error = %v, want nil", nIdx, err)
			}
			n := neigh2.Site()

			angle := computeAngleCCW(c, n, center)
			if angle <= 0 {
				t.Errorf("vd.Cell(%d) Neighbors %d,%d not sorted in CCW", i,
					cIdx, nIdx)
			}
		}
	}
}

func TestTriangleCircumcenter(t *testing.T) {
	tests := []struct {
		name       string
		p0, p1, p2 s2.Point
		want       s2.Point
	}{
		{
			"xyz orthonormal",
			s2.PointFromCoords(1, 0, 0),
			s2.PointFromCoords(0, 1, 0),
			s2.PointFromCoords(0, 0, 1),
			s2.PointFromCoords(1, 1, 1),
		},
		{
			"xyz orthonormal reversed",
			s2.PointFromCoords(0, 0, 1),
			s2.PointFromCoords(0, 1, 0),
			s2.PointFromCoords(1, 0, 0),
			s2.PointFromCoords(1, 1, 1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := triangleCircumcenter(tt.p0, tt.p1, tt.p2)
			if got.Distance(tt.want) > 1e-9 {
				t.Errorf("triangleCircumcenter(...) = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmarks

func BenchmarkNewDiagram(b *testing.B) {
	sizes := []int{1e+2, 1e+3, 1e+4, 1e+5}
	for _, pointsCnt := range sizes {
		b.Run(fmt.Sprintf("N%d", pointsCnt), func(b *testing.B) {
			points := utils.GenerateRandomPoints(pointsCnt, 0)

			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				_, err := NewDiagram(points)
				if err != nil {
					b.Fatalf("NewDiagram(...) error = %v, want nil", err)
				}
			}
		})
	}
}

// Helpers

func mustNewDiagram(t *testing.T, n int) *Diagram {
	t.Helper()
	points := utils.GenerateRandomPoints(n, 0)
	vd, err := NewDiagram(points)
	if err != nil {
		t.Fatalf("NewDiagram(...) error = %v, want nil", err)
	}
	return vd
}

func computeAngleCCW(refVec, vec, normal s2.Point) float64 {
	cross := refVec.Cross(vec.Vector)
	angle := math.Atan2(
		math.Copysign(cross.Norm(), cross.Dot(normal.Vector)),
		refVec.Dot(vec.Vector),
	)
	if angle < 0 {
		angle += 2 * math.Pi
	}
	return angle
}
