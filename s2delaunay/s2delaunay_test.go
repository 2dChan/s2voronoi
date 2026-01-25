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

// Triangle

func TestTrianglePrevVertex(t *testing.T) {
	verts := [3]int{1, 2, 3}
	tri := Triangle{V: verts}
	for i, in := range tri.V {
		got := tri.PrevVertex(in)
		want := verts[(i+2)%len(tri.V)]
		if got != want {
			t.Errorf("tri.PrevVertex(%v) = %v, want %v", in, got, want)
		}
	}
}

func TestTrianglePrevVertex_Panic(t *testing.T) {
	tri := Triangle{V: [3]int{1, 2, 3}}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("PrevVertex should panic for vIdx not in triangle")
		}
	}()
	tri.PrevVertex(-1)
}

func TestTriangleNextVertex(t *testing.T) {
	verts := [3]int{1, 2, 3}
	tri := Triangle{V: verts}
	for i, in := range tri.V {
		got := tri.NextVertex(in)
		want := verts[(i+1)%len(tri.V)]
		if got != want {
			t.Errorf("tri.NextVertex(%v) = %v, want %v", in, got, want)
		}
	}
}

func TestTriangleNextVertex_Panic(t *testing.T) {
	tri := Triangle{V: [3]int{1, 2, 3}}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("PrevVertex should panic for vIdx not in triangle")
		}
	}()
	tri.NextVertex(-1)
}

// DelaunayTriangulation

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

func BenchmarkComputeDelaunayTriangulation(b *testing.B) {
	sizes := []int{1e+2, 1e+3, 1e+4, 1e+5}
	for _, pointsCnt := range sizes {
		b.Run(fmt.Sprintf("N%d", pointsCnt), func(b *testing.B) {
			points := utils.GenerateRandomPoints(pointsCnt, 0)

			b.ResetTimer()
			for b.Loop() {
				_, err := ComputeDelaunayTriangulation(points, 0)
				if err != nil {
					b.Fatalf("ComputeDelaunayTriangulation(...) error = %v, want nil", err)
				}
			}
		})
	}
}

func TestIncidentTriangles(t *testing.T) {
	dt := &DelaunayTriangulation{
		Vertices:                nil,
		Triangles:               nil,
		IncidentTriangleIndices: []int{0, 1, 1, 1, 2},
		IncidentTriangleOffsets: []int{0, 2, 3, 5},
	}

	tests := []struct {
		name    string
		in      int
		want    []int
		wantErr bool
	}{
		{"negative", -1, nil, true},
		{"out of range", len(dt.IncidentTriangleOffsets), nil, true},
		{"index 0", 0, []int{0, 1}, false},
		{"index 1", 1, []int{1}, false},
		{"index 2", 2, []int{1, 2}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dt.IncidentTriangles(tt.in)
			if tt.wantErr && err == nil {
				t.Errorf("dt.IncidentTriangles(%d) error = nil, want non-nil", tt.in)
			}
			if tt.wantErr == false && err != nil {
				t.Errorf("dt.IncidentTriangles(%d) error = %v, want nil", tt.in, err)
			}
			if tt.wantErr == false && err == nil && cmp.Equal(tt.want, got) == false {
				t.Errorf("dt.IncidentTriangles(%d) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestTriangleVertices(t *testing.T) {
	points := utils.GenerateRandomPoints(3, 0)
	dt := &DelaunayTriangulation{
		Vertices: s2.PointVector{points[0], points[1], points[2]},
		Triangles: []Triangle{
			{V: [3]int{0, 1, 2}},
		},
	}

	want := [3]s2.Point{points[0], points[1], points[2]}
	if got, err := dt.TriangleVertices(0); err != nil {
		t.Errorf("dt.TriangleVertices(0) error = %v, want nil", err)
	} else if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("dt.TriangleVertices(0) mismatch (-want +got):\n%v", diff)
	}

	if _, err := dt.TriangleVertices(1); err == nil {
		t.Error("dt.TriangleVertices(1) error = got nil, want non-nil")
	}
	if _, err := dt.TriangleVertices(-1); err == nil {
		t.Error("dt.TriangleVertices(-1) error = got nil, want non-nil")
	}
}

func TestComputeDelaunayTriangulation_DegenerateInput(t *testing.T) {
	vertices := s2.PointVector{
		s2.PointFromCoords(1, 0, 0),
		s2.PointFromCoords(0, 1, 0),
		s2.PointFromCoords(0, 0, 1),
	}
	if _, err := ComputeDelaunayTriangulation(vertices, 0); err == nil {
		t.Errorf("ComputeDelaunayTriangulation(...) error = nil, want non-nil")
	}
}

func TestComputeDelaunayTriangulation_VerticesOnSphere(t *testing.T) {
	dt := mustComputeDelaunayTriangulation(t, 100)

	for i, p := range dt.Vertices {
		norm := p.Norm()
		if math.Abs(norm-1.0) > defaultEps {
			t.Errorf(
				"ComputeDelaunayTriangulation(...).Vertices[%d] norm = %v, want ~1.0", i,
				norm)
		}
	}
}

func TestComputeDelaunayTriangulation_VerifyTrianglesCCW(t *testing.T) {
	dt := mustComputeDelaunayTriangulation(t, 100)

	for i, tt := range dt.Triangles {
		p0, p1, p2 := dt.Vertices[tt.V[0]], dt.Vertices[tt.V[1]], dt.Vertices[tt.V[2]]
		cross := p1.Sub(p0.Vector).Cross(p2.Sub(p0.Vector))
		dot := cross.Dot(p0.Vector)
		if dot < 0 {
			t.Errorf("ComputeDelaunayTriangulation(...).Triangles[%d] vertices not sorted in CCW",
				i)
		}
	}
}

func TestComputeDelaunayTriangulation_VerifyIncidentTrianglesSorted(t *testing.T) {
	dt := mustComputeDelaunayTriangulation(t, 100)

	for vIdx := range len(dt.Vertices) {
		incidentTris, err := dt.IncidentTriangles(vIdx)
		if err != nil {
			t.Fatalf("dt.IncidentTriangles(%v) error = %v", vIdx, err)
		}
		for i := 1; i < len(incidentTris); i++ {
			ct := dt.Triangles[incidentTris[i-1]]
			nt := dt.Triangles[incidentTris[i]]

			nextVertex := ct.NextVertex(vIdx)
			prevVertex := nt.PrevVertex(vIdx)

			if nextVertex != prevVertex {
				t.Errorf("dt.IncidentTriangles(%d) triangles %v and %v not CCW neighbors", vIdx,
					i-1, i)
			}
		}
	}
}

func TestSortTriangleVerticesCCW(t *testing.T) {
	a := s2.PointFromCoords(1, 0, 0)
	b := s2.PointFromCoords(0, 1, 0)
	c := s2.PointFromCoords(0, 0, 1)
	verts := s2.PointVector{a, b, c}

	want1 := [3]int{0, 1, 2}
	tri1 := &Triangle{V: [3]int{0, 1, 2}}
	sortTriangleVerticesCCW(tri1, verts)
	if tri1.V != want1 {
		t.Errorf("sortTriangleVerticesCCW([0 1 2], verts): tri1.V = %v, want %v", tri1.V, want1)
	}

	want2 := [3]int{0, 1, 2}
	tri2 := &Triangle{V: [3]int{0, 2, 1}}
	sortTriangleVerticesCCW(tri2, verts)
	if tri2.V != want2 {
		t.Errorf("sortTriangleVerticesCCW([0 2 1], verts): tri2.V = %v, want %v", tri2.V, want2)
	}
}

func TestSortIncidentTriangleIndicesCCW(t *testing.T) {
	expected3 := []int{0, 2, 1}
	incident3 := []int{0, 1, 2}
	tris3 := []Triangle{
		{V: [3]int{0, 1, 2}},
		{V: [3]int{0, 2, 3}},
		{V: [3]int{0, 3, 1}},
	}
	sortIncidentTriangleIndicesCCW(0, incident3, tris3)
	if cyclicEqual(incident3, expected3) == false {
		t.Errorf("sortIncidentTriangleIndicesCCW(...): incident3 = %v, want %v", incident3, expected3)
	}

	expected4 := []int{1, 0, 3, 2}
	incident4 := []int{1, 3, 2, 0}
	tris4 := []Triangle{
		{V: [3]int{0, 1, 2}},
		{V: [3]int{0, 2, 3}},
		{V: [3]int{0, 3, 4}},
		{V: [3]int{0, 4, 1}},
	}
	sortIncidentTriangleIndicesCCW(0, incident4, tris4)
	if cyclicEqual(incident4, expected4) == false {
		t.Errorf("sortIncidentTriangleIndicesCCW(...): incident4 = %v, want %v", incident4, expected4)
	}
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

func mustComputeDelaunayTriangulation(t *testing.T, n int) *DelaunayTriangulation {
	t.Helper()
	vertices := utils.GenerateRandomPoints(n, 0)

	dt, err := ComputeDelaunayTriangulation(vertices, 0)
	if err != nil {
		t.Fatalf("ComputeDelaunayTriangulation(...) error = %v, want nil", err)
	}
	return dt
}
