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
	"github.com/google/go-cmp/cmp"
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
			vd, err := NewDiagram(points, WithEps(tt.eps))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDiagram(..., WithEps(%v)) error = %v, wantErr %v", tt.eps, err, tt.wantErr)
			}

			if err == nil && vd.eps != tt.eps {
				t.Errorf("NewDiagram(..., WithEps(%v)) eps = %v, want %v", tt.eps, vd.eps, tt.eps)
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
		cell := vd.Cell(i)

		center := cell.Site()
		for i := 0; i < cell.NumVertices(); i++ {
			cIdx := i
			c := cell.Vertex(cIdx)
			nIdx := (i + 1) % cell.NumVertices()
			n := cell.Vertex(nIdx)

			angle := computeAngleCCW(c, n, center)
			if angle <= 0 {
				t.Errorf("vd.Cell(%d) Vertices %d,%d not sorted in CCW", i,
					cIdx, nIdx)
			}
		}

		for i := 0; i < cell.NumNeighbors(); i++ {
			cIdx := i
			cn := cell.Neighbor(cIdx)
			c := cn.Site()

			nIdx := (i + 1) % cell.NumNeighbors()
			nn := cell.Neighbor(nIdx)
			n := nn.Site()

			angle := computeAngleCCW(c, n, center)
			if angle <= 0 {
				t.Errorf("vd.Cell(%d) Neighbors %d,%d not sorted in CCW", i, cIdx, nIdx)
			}
		}
	}
}

func TestDiagram_NumCells(t *testing.T) {
	vd := mustNewDiagram(t, 10)
	want := len(vd.Sites)
	got := vd.NumCells()
	if got != want {
		t.Errorf("Diagram.NumCells() = %d, want %d", got, want)
	}
}

func TestDiagram_Cell(t *testing.T) {
	vd := mustNewDiagram(t, 10)
	for i := range vd.NumCells() {
		c := vd.Cell(i)
		want := Cell{i, vd}
		if diff := cmp.Diff(want, c, cmp.AllowUnexported(Cell{}, Diagram{})); diff != "" {
			t.Errorf("vd.Cell(%d) mismatch (-want +got):\n%s", i, diff)
		}
	}
}

func TestDiagram_Cell_Panic(t *testing.T) {
	vd := mustNewDiagram(t, 10)

	tests := []struct {
		name  string
		index int
	}{
		{"negative index", -1},
		{"out of range", vd.NumCells()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("vd.Cell(%d) did not panic, want panic", tt.index)
				}
			}()
			vd.Cell(tt.index)
		})
	}
}

func TestDiagram_Relax(t *testing.T) {
	tests := []struct {
		name  string
		steps int
		size  int
	}{
		{"zero step", 0, 1000},
		{"one step", 1, 1000},
		{"multiple steps", 5, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vd := mustNewDiagram(t, tt.size)
			vdOld := mustNewDiagram(t, tt.size)

			err := vd.Relax(tt.steps)
			if err != nil {
				t.Fatalf("vd.Relax(%d) error = %v, want nil", tt.steps, err)
			}

			if len(vd.Sites) != len(vdOld.Sites) {
				t.Errorf("vd.Relax(%d) Sites count = %d, want %d", tt.steps,
					len(vd.Sites), len(vdOld.Sites))
			}
			if len(vd.Vertices) != len(vdOld.Vertices) {
				t.Errorf("vd.Relax(%d) Vertices count = %d, want %d", tt.steps,
					len(vd.Vertices), len(vdOld.Vertices))
			}
			if len(vd.CellNeighbors) != len(vdOld.CellNeighbors) {
				t.Errorf("vd.Relax(%d) CellNeighbors count = %d, want %d", tt.steps,
					len(vd.CellNeighbors), len(vdOld.CellNeighbors))
			}
			if len(vd.CellVertices) != len(vdOld.CellVertices) {
				t.Errorf("vd.Relax(%d) CellVertices count = %d, want %d", tt.steps,
					len(vd.CellVertices), len(vdOld.CellVertices))
			}
			if len(vd.CellOffsets) != len(vdOld.CellOffsets) {
				t.Errorf("vd.Relax(%d) CellOffsets count = %d, want %d", tt.steps,
					len(vd.CellOffsets), len(vdOld.CellOffsets))
			}

			expectChange := tt.steps != 0
			msg := "changed"
			if expectChange {
				msg = "not changed"
			}
			if cmp.Equal(vd.Sites, vdOld.Sites) == expectChange {
				t.Errorf("vd.Relax(%d) Sites %s", tt.steps, msg)
			}
			if cmp.Equal(vd.Vertices, vdOld.Vertices) == expectChange {
				t.Errorf("vd.Relax(%d) Vertices %s", tt.steps, msg)
			}
			if cmp.Equal(vd.CellNeighbors, vdOld.CellNeighbors) == expectChange {
				t.Errorf("vd.Relax(%d) CellNeighbors %s", tt.steps, msg)
			}
			if cmp.Equal(vd.CellVertices, vdOld.CellVertices) == expectChange {
				t.Errorf("vd.Relax(%d) CellVertices %s", tt.steps, msg)
			}
			if cmp.Equal(vd.CellOffsets, vdOld.CellOffsets) == expectChange {
				t.Errorf("vd.Relax(%d) CellOffsets %s", tt.steps, msg)
			}
		})
	}

	vd := mustNewDiagram(t, 100)
	if err := vd.Relax(-1); err == nil {
		t.Errorf("vd.Relax(-1) error = nil, want non-nil")
	}
}

func TestTriangleCircumcenter(t *testing.T) {
	tests := []struct {
		name    string
		a, b, c s2.Point
		want    s2.Point
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
			got := triangleCircumcenter(tt.a, tt.b, tt.c)
			if got.Distance(tt.want) > defaultEps {
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

func BenchmarkDiagram_Relax(b *testing.B) {
	sizes := []int{1e+2, 1e+3, 1e+4}
	steps := []int{1, 10}
	for _, pointsCnt := range sizes {
		for _, step := range steps {
			b.Run(fmt.Sprintf("N%d Steps%d", pointsCnt, step), func(b *testing.B) {
				points := utils.GenerateRandomPoints(pointsCnt, 0)

				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					b.StopTimer()
					vd, err := NewDiagram(points)
					if err != nil {
						b.Fatalf("NewDiagram(...) error = %v, want nil", err)
					}
					b.StartTimer()

					err = vd.Relax(step)
					if err != nil {
						b.Fatalf("vd.Relax(%d) error = %v, want nil", step, err)
					}
				}
			})
		}
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
