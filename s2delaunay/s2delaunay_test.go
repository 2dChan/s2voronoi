// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

package s2delaunay

import (
	"fmt"
	"math"
	"testing"

	"github.com/2dChan/s2voronoi/utils"
	"github.com/golang/geo/r3"
	"github.com/golang/geo/s2"
	"github.com/google/go-cmp/cmp"
	"github.com/markus-wa/quickhull-go/v2"
)

// TriangulationOptions

func TestWithEps(t *testing.T) {
	const (
		eps = 0.5
	)

	opts := &TriangulationOptions{Eps: 0}
	opt := WithEps(eps)
	opt(opts)
	if opts.Eps != eps {
		t.Errorf("WithEps(%v): opts.Eps = %v, want %v", eps, opts.Eps, eps)
	}
}

func TestWithEps_Panic(t *testing.T) {
	invalidEps := []float64{-1.0, -0.1, 0.0}
	for _, eps := range invalidEps {
		t.Run(fmt.Sprintf("eps %v", eps), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("WithEps(%v): shoud panic for eps<=0", eps)
				}
			}()
			WithEps(eps)
		})
	}
}

// Triangulation

func TestNewTriangulation_WithEps(t *testing.T) {
	const (
		customEps = 0.01
	)
	vertices := utils.GenerateRandomPoints(10, 0)
	_, err := NewTriangulation(vertices, WithEps(customEps))
	if err != nil {
		t.Fatalf("NewTriangulation(...): error = %v, want nil", err)
	}
}

func TestNewTriangulation_DegenerateInput(t *testing.T) {
	vertices := s2.PointVector{
		s2.PointFromCoords(1, 0, 0),
		s2.PointFromCoords(0, 1, 0),
		s2.PointFromCoords(0, 0, 1),
	}
	if _, err := NewTriangulation(vertices); err == nil {
		t.Errorf("NewTriangulation(...) error = nil, want non-nil")
	}
}

func TestNewTriangulation_VerticesOnSphere(t *testing.T) {
	dt := mustNewTriangulation(t, 100)

	for i, p := range dt.Vertices {
		norm := p.Norm()
		if math.Abs(norm-1.0) > defaultEps {
			t.Errorf(
				"NewTriangulation(...).Vertices[%d] norm = %v, want ~1.0", i,
				norm)
		}
	}
}

func TestNewTriangulation_VerifyTrianglesCCW(t *testing.T) {
	dt := mustNewTriangulation(t, 100)

	for i, tri := range dt.Triangles {
		p0, p1, p2 := dt.Vertices[tri[0]], dt.Vertices[tri[1]], dt.Vertices[tri[2]]
		cross := p1.Sub(p0.Vector).Cross(p2.Sub(p0.Vector))
		dot := cross.Dot(p0.Vector)
		if dot < 0 {
			t.Errorf("NewTriangulation(...).Triangles[%d] vertices not sorted in CCW",
				i)
		}
	}
}

func TestNewTriangulation_VerifyIncidentTrianglesSorted(t *testing.T) {
	dt := mustNewTriangulation(t, 100)

	for vIdx := range len(dt.Vertices) {
		incidentTris := dt.IncidentTriangles(vIdx)
		for i := 1; i < len(incidentTris); i++ {
			ct := dt.Triangles[incidentTris[i-1]]
			nt := dt.Triangles[incidentTris[i]]

			nextVertex := NextVertex(ct, vIdx)
			prevVertex := PrevVertex(nt, vIdx)

			if nextVertex != prevVertex {
				t.Errorf("dt.IncidentTriangles(%d) triangles %v and %v not CCW neighbors", vIdx,
					i-1, i)
			}
		}
	}
}

func TestIncidentTriangles(t *testing.T) {
	dt := &Triangulation{
		Vertices:                nil,
		Triangles:               nil,
		IncidentTriangleIndices: []int{0, 1, 1, 1, 2},
		IncidentTriangleOffsets: []int{0, 2, 3, 5},
	}

	tests := []struct {
		name string
		in   int
		want []int
	}{
		{"index 0", 0, []int{0, 1}},
		{"index 1", 1, []int{1}},
		{"index 2", 2, []int{1, 2}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dt.IncidentTriangles(tt.in)
			if cmp.Equal(tt.want, got) == false {
				t.Errorf("dt.IncidentTriangles(%d) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
func TestIncidentTriangles_Panic(t *testing.T) {
	dt := &Triangulation{
		Vertices:                nil,
		Triangles:               nil,
		IncidentTriangleIndices: []int{0, 1, 1, 1, 2},
		IncidentTriangleOffsets: []int{0, 2, 3, 5},
	}

	checkPanic := func(idx int) {
		panicked := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			_ = dt.IncidentTriangles(idx)
		}()
		if panicked == false {
			t.Errorf("dt.IncidentTriangles(%d) did not panic, want panic", idx)
		}
	}

	checkPanic(-1)
	checkPanic(len(dt.IncidentTriangleOffsets))
}

func TestTriangleVertices(t *testing.T) {
	points := utils.GenerateRandomPoints(3, 0)
	dt := &Triangulation{
		Vertices: s2.PointVector{points[0], points[1], points[2]},
		Triangles: [][3]int{
			{0, 1, 2},
		},
	}

	want := [3]s2.Point{points[0], points[1], points[2]}
	p0, p1, p2 := dt.TriangleVertices(0)
	got := [3]s2.Point{p0, p1, p2}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("dt.TriangleVertices(0) mismatch (-want +got):\n%v", diff)
	}
}

func TestTriangleVertices_Panic(t *testing.T) {
	points := utils.GenerateRandomPoints(3, 0)
	dt := &Triangulation{
		Vertices: s2.PointVector{points[0], points[1], points[2]},
		Triangles: [][3]int{
			{0, 1, 2},
		},
	}

	checkPanic := func(idx int) {
		panicked := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					panicked = true
				}
			}()
			dt.TriangleVertices(idx)
		}()
		if panicked == false {
			t.Errorf("dt.TriangleVertices(%d) did not panic, want panic", idx)
		}
	}

	checkPanic(1)
	checkPanic(-1)
}

func TestSortTriangleVerticesCCW(t *testing.T) {
	a := s2.PointFromCoords(1, 0, 0)
	b := s2.PointFromCoords(0, 1, 0)
	c := s2.PointFromCoords(0, 0, 1)
	verts := s2.PointVector{a, b, c}

	want1 := [3]int{0, 1, 2}
	tri1 := [3]int{0, 1, 2}
	sortTriangleVerticesCCW(&tri1, verts)
	if diff := cmp.Diff(want1, tri1); diff != "" {
		t.Errorf("sortTriangleVerticesCCW([0 1 2], verts) mismatch (-want +got):\n%s", diff)
	}

	want2 := [3]int{0, 1, 2}
	tri2 := [3]int{0, 2, 1}
	sortTriangleVerticesCCW(&tri2, verts)
	if diff := cmp.Diff(want2, tri2); diff != "" {
		t.Errorf("sortTriangleVerticesCCW([0 2 1], verts) mismatch (-want +got):\n%s", diff)
	}
}

func TestSortIncidentTriangleIndicesCCW(t *testing.T) {
	expected3 := []int{0, 2, 1}
	incident3 := []int{0, 1, 2}
	tris3 := [][3]int{
		{0, 1, 2},
		{0, 2, 3},
		{0, 3, 1},
	}
	sortIncidentTriangleIndicesCCW(0, incident3, tris3)
	if cyclicEqual(incident3, expected3) == false {
		t.Errorf("sortIncidentTriangleIndicesCCW(...): incident3 = %v, want %v", incident3, expected3)
	}

	expected4 := []int{1, 0, 3, 2}
	incident4 := []int{1, 3, 2, 0}
	tris4 := [][3]int{
		{0, 1, 2},
		{0, 2, 3},
		{0, 3, 4},
		{0, 4, 1},
	}
	sortIncidentTriangleIndicesCCW(0, incident4, tris4[:])
	if cyclicEqual(incident4, expected4) == false {
		t.Errorf("sortIncidentTriangleIndicesCCW(...): incident4 = %v, want %v", incident4, expected4)
	}
}

// Triangle Prev/Next vertex

func TestPrevVertex(t *testing.T) {
	tri := [3]int{1, 2, 3}
	for i, in := range tri {
		got := PrevVertex(tri, in)
		want := tri[(i+2)%len(tri)]
		if got != want {
			t.Errorf("tri.PrevVertex(%v) = %v, want %v", in, got, want)
		}
	}
}

func TestPrevVertex_Panic(t *testing.T) {
	tri := [3]int{1, 2, 3}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("PrevVertex should panic for vIdx not in triangle")
		}
	}()
	PrevVertex(tri, -1)
}

func TestNextVertex(t *testing.T) {
	tri := [3]int{1, 2, 3}
	for i, in := range tri {
		got := NextVertex(tri, in)
		want := tri[(i+1)%len(tri)]
		if got != want {
			t.Errorf("tri.NextVertex(%v) = %v, want %v", in, got, want)
		}
	}
}

func TestNextVertex_Panic(t *testing.T) {
	tri := [3]int{1, 2, 3}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("PrevVertex should panic for vIdx not in triangle")
		}
	}()
	NextVertex(tri, -1)
}

// Benchmarks

func BenchmarkConvexHull(b *testing.B) {
	sizes := []int{1e+2, 1e+3, 1e+4, 1e+5}
	for _, pointsCnt := range sizes {
		b.Run(fmt.Sprintf("N%d", pointsCnt), func(b *testing.B) {
			points := utils.GenerateRandomPoints(pointsCnt, 0)
			v3 := make([]r3.Vector, len(points))
			for i, p := range points {
				v3[i] = p.Vector
			}

			qh := new(quickhull.QuickHull)

			b.ResetTimer()
			for b.Loop() {
				qh.ConvexHull(v3, true, true, 0)
			}
		})
	}
}

func BenchmarkNewTriangulation(b *testing.B) {
	sizes := []int{1e+2, 1e+3, 1e+4, 1e+5}
	for _, pointsCnt := range sizes {
		b.Run(fmt.Sprintf("N%d", pointsCnt), func(b *testing.B) {
			points := utils.GenerateRandomPoints(pointsCnt, 0)

			b.ResetTimer()
			for b.Loop() {
				_, err := NewTriangulation(points)
				if err != nil {
					b.Fatalf("NewTriangulation(...) error = %v, want nil", err)
				}
			}
		})
	}
}

// Helpers

func mustNewTriangulation(t *testing.T, n int) *Triangulation {
	t.Helper()
	vertices := utils.GenerateRandomPoints(n, 0)

	dt, err := NewTriangulation(vertices)
	if err != nil {
		t.Fatalf("NewTriangulation(...) error = %v, want nil", err)
	}
	return dt
}

func cyclicEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	n := len(a)
	for i := range n {
		if b[0] != a[i] {
			continue
		}

		equal := true
		for j := range n {
			if a[(i+j)%n] != b[j] {
				equal = false
				break
			}
		}
		if equal {
			return true
		}
	}

	return false
}
