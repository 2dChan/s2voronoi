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
	const (
		eps = 0.5
	)

	opts := &DiagramOptions{Eps: 0}
	opt := WithEps(eps)
	opt(opts)
	if opts.Eps != eps {
		t.Errorf("WithEps(%v) opts.Eps = %v, want %v", eps, opts.Eps, eps)
	}
}

func TestWithEps_Panic(t *testing.T) {
	invalidEps := []float64{-1.0, -0.1, 0.0}
	for _, eps := range invalidEps {
		t.Run(fmt.Sprintf("eps %v", eps), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("WithEps(%v) should panic for eps<=0", eps)
				}
			}()
			WithEps(eps)
		})
	}
}

// Diagram

func TestNewDiagram_WithEps(t *testing.T) {
	const (
		customEps = 0.01
	)
	vertices := utils.GenerateRandomPoints(10, 0)
	_, err := NewDiagram(vertices, WithEps(customEps))
	if err != nil {
		t.Fatalf("NewDiagram(...): error = %v, want nil", err)
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
		cell := vd.Cell(i)

		center := cell.Site()
		for i := 0; i < cell.NumVertices(); i++ {
			cIdx := i
			nIdx := (i + 1) % cell.NumVertices()
			c := cell.Vertex(cIdx)
			n := cell.Vertex(nIdx)

			angle := computeAngleCCW(c, n, center)
			if angle <= 0 {
				t.Errorf("vd.Cell(%d) Vertices %d,%d not sorted in CCW", i,
					cIdx, nIdx)
			}
		}

		for i := 0; i < cell.NumNeighbors(); i++ {
			cIdx := i
			nIdx := (i + 1) % cell.NumNeighbors()
			c := cell.Neighbor(cIdx).Site()
			n := cell.Neighbor(nIdx).Site()

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
