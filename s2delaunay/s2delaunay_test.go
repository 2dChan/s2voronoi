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
			opts := &TriangulationOptions{Eps: defaultEps}
			opt := WithEps(tt.eps)
			err := opt(opts)
			if (err != nil) != tt.wantErr {
				errValMsg := "nil"
				if tt.wantErr {
					errValMsg = "non-nil"
				}
				t.Errorf("WithEps(%v) error = %v, want %v", tt.eps, err, errValMsg)
			}
			if err == nil && opts.Eps != tt.eps {
				t.Errorf("WithEps(%v) opts.Eps = %v, want %v", tt.eps, opts.Eps, tt.eps)
			}
		})
	}
}

// Triangulation

func TestNewTriangulation_WithEps(t *testing.T) {
	points := utils.GenerateRandomPoints(10, 0)
	tests := []struct {
		name    string
		eps     float64
		wantErr bool
	}{
		{"eps default", defaultEps, false},
		{"eps positive", 0.01, false},
		{"eps large", 1, true},
		{"eps zero", 0, true},
		{"eps negative", -0.01, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTriangulation(points, WithEps(tt.eps))
			if (err != nil) != tt.wantErr {
				errValMsg := "nil"
				if tt.wantErr {
					errValMsg = "non-nil"
				}
				t.Errorf("NewTriangulation(..., WithEps(%v)) error = %v, want %s", tt.eps, err, errValMsg)
			}
		})
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
				"dt.Vertices[%d] norm = %v, want ~1.0", i,
				norm)
		}
	}
}

func TestNewTriangulation_VerifyTrianglesCCW(t *testing.T) {
	dt := mustNewTriangulation(t, 100)

	for i, tri := range dt.Triangles {
		a, b, c := dt.Vertices[tri[0]], dt.Vertices[tri[1]], dt.Vertices[tri[2]]
		cross := b.Sub(a.Vector).Cross(c.Sub(a.Vector))
		dot := cross.Dot(a.Vector)
		if dot < 0 {
			t.Errorf("dt.Triangles[%d] vertices are not sorted in CCW", i)
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
				t.Errorf("dt.IncidentTriangles(%d) triangles %d and %d are not CCW neighbors", vIdx, i-1, i)
			}
		}
	}
}

func TestTriangulation_IncidentTriangles(t *testing.T) {
	assertPanic := func(dt *Triangulation, in int) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("dt.IncidentTriangles(%d) did not panic, want panic", in)
			}
		}()
		dt.IncidentTriangles(in)
	}

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

	assertPanic(dt, -1)
	assertPanic(dt, len(dt.IncidentTriangleOffsets))

}

func TestTriangulation_TriangleVertices(t *testing.T) {
	assertPanic := func(dt *Triangulation, in int) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("dt.TriangleVertices(%d) did not panic, want panic", in)
			}
		}()
		dt.TriangleVertices(in)
	}

	points := utils.GenerateRandomPoints(3, 0)
	dt := &Triangulation{
		Vertices: s2.PointVector{points[0], points[1], points[2]},
		Triangles: [][3]int{
			{0, 1, 2},
		},
	}

	want := [3]s2.Point{points[0], points[1], points[2]}
	a, b, c := dt.TriangleVertices(0)
	got := [3]s2.Point{a, b, c}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("dt.TriangleVertices(0) mismatch (-want +got):\n%s", diff)
	}

	assertPanic(dt, -1)
	assertPanic(dt, len(dt.Triangles))

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
		t.Errorf("sortIncidentTriangleIndicesCCW(...) incident3 = %v, want %v", incident3, expected3)
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
		t.Errorf("sortIncidentTriangleIndicesCCW(...) incident4 = %v, want %v", incident4, expected4)
	}
}

// Triangle Prev/Next vertex

func TestPrevVertex(t *testing.T) {
	assertPanic := func(tri [3]int, in int) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("PrevVertex(%v, %d) did not panic, want panic", tri, in)
			}
		}()
		PrevVertex(tri, in)
	}

	tri := [3]int{1, 2, 3}
	for i, in := range tri {
		got := PrevVertex(tri, in)
		want := tri[(i+2)%len(tri)]
		if got != want {
			t.Errorf("PrevVertex(%v, %d) = %v, want %v", tri, in, got, want)
		}
	}

	assertPanic(tri, -1)
	assertPanic(tri, 4)
}

func TestNextVertex(t *testing.T) {
	assertPanic := func(tri [3]int, in int) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("NextVertex(%v, %d) did not panic, want panic", tri, in)
			}
		}()
		NextVertex(tri, in)
	}

	tri := [3]int{1, 2, 3}
	for i, in := range tri {
		got := NextVertex(tri, in)
		want := tri[(i+1)%len(tri)]
		if got != want {
			t.Errorf("NextVertex(%v, %d) = %v, want %v", tri, in, got, want)
		}
	}

	assertPanic(tri, -1)
	assertPanic(tri, 4)
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

			b.ReportAllocs()
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

			b.ReportAllocs()
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
