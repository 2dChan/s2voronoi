// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

// Package s2delaunay implements Delaunay triangulation on the S2 sphere using convex hull algorithms.

package s2delaunay

import (
	"errors"
	"fmt"

	"github.com/golang/geo/r3"
	"github.com/golang/geo/s2"
	"github.com/markus-wa/quickhull-go/v2"
)

const (
	defaultEps = 1e-12
)

// Triangulation represents a Delaunay triangulation on the S2 sphere.
type Triangulation struct {
	// Vertices are the input points on the unit sphere.
	Vertices s2.PointVector
	// Triangles are the triangulation triangles, each with three vertex indices,
	// sorted CCW when looking out of the sphere.
	Triangles [][3]int
	// IncidentTriangleIndices contains indices of incident triangles for each vertex,
	// sorted CCW when looking out of the sphere, forming a CSR-like sparse representation.
	IncidentTriangleIndices []int
	// IncidentTriangleOffsets contains offsets for slicing incident triangle data in a CSR-like format.
	IncidentTriangleOffsets []int
}

// TriangulationOptions holds configuration options for Delaunay triangulation.
type TriangulationOptions struct {
	Eps float64
}

// TriangulationOption is a functional option type for triangulation configuration.
type TriangulationOption func(*TriangulationOptions) error

// WithEps sets the numerical precision epsilon for triangulation.
// It must be positive.
func WithEps(eps float64) TriangulationOption {
	return func(o *TriangulationOptions) error {
		if eps <= 0 {
			return fmt.Errorf("WithEps: eps must be positive got %v", eps)
		}
		o.Eps = eps
		return nil
	}
}

// NewTriangulation creates a Delaunay triangulation from the given vertices.
// The vertices must lie on the unit sphere, there must be at least 4 vertices, and they must not be coplanar.
// It returns an error if the triangulation cannot be constructed.
func NewTriangulation(vertices s2.PointVector, setters ...TriangulationOption) (*Triangulation,
	error) {
	opts := TriangulationOptions{
		Eps: defaultEps,
	}
	for _, set := range setters {
		err := set(&opts)
		if err != nil {
			return nil, err
		}
	}
	numVertices := len(vertices)
	if numVertices < 4 {
		return nil,
			errors.New("NewTriangulation: insufficient vertices for triangulation minimum 4 required")
	}
	numTriangles := 2 * (numVertices - 2)
	t := &Triangulation{
		Vertices:                vertices,
		Triangles:               make([][3]int, numTriangles),
		IncidentTriangleIndices: make([]int, numTriangles*3),
		IncidentTriangleOffsets: make([]int, numVertices+1),
	}
	r3vertices := make([]r3.Vector, numVertices)
	for i, p := range vertices {
		r3vertices[i] = p.Vector
	}
	qh := new(quickhull.QuickHull)
	ch := qh.ConvexHull(r3vertices, true, true, opts.Eps)
	if len(ch.Indices) != numTriangles*3 {
		return nil,
			errors.New("NewTriangulation: inconsistent number of indices returned from QuickHull")
	}
	for _, idx := range ch.Indices {
		t.IncidentTriangleOffsets[idx+1]++
	}
	for i := range numVertices {
		t.IncidentTriangleOffsets[i+1] += t.IncidentTriangleOffsets[i]
	}
	nxt := make([]int, numVertices)
	copy(nxt, t.IncidentTriangleOffsets[:numVertices])
	for i := range numTriangles {
		base := i * 3
		for j := range 3 {
			v := ch.Indices[base+j]
			t.Triangles[i][j] = v
			t.IncidentTriangleIndices[nxt[v]] = i
			nxt[v]++
		}
		sortTriangleVerticesCCW(&t.Triangles[i], t.Vertices)
	}
	for i := range numVertices {
		incidentTriangles := t.IncidentTriangles(i)
		sortIncidentTriangleIndicesCCW(i, incidentTriangles, t.Triangles)
	}
	return t, nil
}

// IncidentTriangles returns the indices of triangles incident to the vertex at the given index,
// sorted in CCW order when looking out of the sphere.
// It panics if the vertex index is out of range.
func (t *Triangulation) IncidentTriangles(vIdx int) []int {
	if vIdx < 0 || vIdx+1 >= len(t.IncidentTriangleOffsets) {
		panic(fmt.Sprintf("IncidentTriangles: vIdx %d out of range [0 %d)", vIdx,
			len(t.IncidentTriangleOffsets)-1))
	}
	start := t.IncidentTriangleOffsets[vIdx]
	end := t.IncidentTriangleOffsets[vIdx+1]
	return t.IncidentTriangleIndices[start:end]
}

// TriangleVertices returns the three vertices of the triangle at the given index.
// It panics if the triangle index is out of bounds.
func (t *Triangulation) TriangleVertices(tIdx int) (s2.Point, s2.Point, s2.Point) {
	if tIdx < 0 || tIdx >= len(t.Triangles) {
		panic(fmt.Sprintf("TriangleVertices: tIdx %d out of bounds [0 %d)", tIdx, len(t.Triangles)))
	}
	tri := t.Triangles[tIdx]
	return t.Vertices[tri[0]], t.Vertices[tri[1]], t.Vertices[tri[2]]
}

// sortTriangleVerticesCCW sorts triangle vertices in CCW order.
func sortTriangleVerticesCCW(t *[3]int, v s2.PointVector) {
	a, b, c := v[t[0]], v[t[1]], v[t[2]]
	norm := b.Sub(a.Vector).Cross(c.Sub(a.Vector))
	if norm.Dot(a.Vector) < 0 {
		t[1], t[2] = t[2], t[1]
	}
}

// sortIncidentTriangleIndicesCCW sorts incident triangle indices in CCW order.
func sortIncidentTriangleIndicesCCW(vIdx int, incidentTris []int, tris [][3]int) {
	n := len(incidentTris)
	for i := 1; i < n; i++ {
		nxt := NextVertex(tris[incidentTris[i-1]], vIdx)
		for j := i + 1; j < n; j++ {
			prv := PrevVertex(tris[incidentTris[j]], vIdx)
			if nxt == prv {
				incidentTris[i], incidentTris[j] = incidentTris[j], incidentTris[i]
				break
			}
		}
	}
}

// PrevVertex returns the previous vertex in the triangle relative to the given vertex index.
// It panics if the vertex index is not part of the triangle.
func PrevVertex(t [3]int, vIdx int) int {
	switch vIdx {
	case t[0]:
		return t[2]
	case t[1]:
		return t[0]
	case t[2]:
		return t[1]
	}
	panic(fmt.Sprintf("PrevVertex: vIdx %d not in triangle %v", vIdx, t))
}

// NextVertex returns the next vertex in the triangle relative to the given vertex index.
// It panics if the vertex index is not part of the triangle.
func NextVertex(t [3]int, vIdx int) int {
	switch vIdx {
	case t[0]:
		return t[1]
	case t[1]:
		return t[2]
	case t[2]:
		return t[0]
	}
	panic(fmt.Sprintf("NextVertex: vIdx %d not in triangle %v", vIdx, t))
}
