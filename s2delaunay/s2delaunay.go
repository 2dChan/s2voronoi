// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

package s2delaunay

import (
	"errors"

	"github.com/golang/geo/r3"
	"github.com/golang/geo/s2"
	"github.com/markus-wa/quickhull-go/v2"
)

const (
	defaultEps = 1e-12
)

type Triangle struct {
	// NOTE: Sort in CCW(look out of sphere)
	V [3]int
}

func (t *Triangle) PrevVertex(vIdx int) int {
	switch vIdx {
	case t.V[0]:
		return t.V[2]
	case t.V[1]:
		return t.V[0]
	case t.V[2]:
		return t.V[1]
	}
	panic("PrevVertex: vIdx not in triangle")
}

func (t *Triangle) NextVertex(vIdx int) int {
	switch vIdx {
	case t.V[0]:
		return t.V[1]
	case t.V[1]:
		return t.V[2]
	case t.V[2]:
		return t.V[0]
	}
	panic("NextVertex: vIdx not in triangle")
}

type DelaunayTriangulation struct {
	Vertices  s2.PointVector
	Triangles []Triangle
	// NOTE: Sort in CCW per vertex(look out of sphere)
	IncidentTriangleIndices []int
	IncidentTriangleOffsets []int
}

func (dt *DelaunayTriangulation) IncidentTriangles(vIdx int) []int {
	if vIdx < 0 || vIdx+1 >= len(dt.IncidentTriangleOffsets) {
		panic("IncidentTriangles: vIdx out of range")
	}
	start := dt.IncidentTriangleOffsets[vIdx]
	end := dt.IncidentTriangleOffsets[vIdx+1]
	return dt.IncidentTriangleIndices[start:end]
}

func (dt *DelaunayTriangulation) TriangleVertices(tIdx int) [3]s2.Point {
	if tIdx < 0 || tIdx >= len(dt.Triangles) {
		panic("TriangleVertices: tIdx out of bounds")
	}
	t := dt.Triangles[tIdx]
	return [3]s2.Point{dt.Vertices[t.V[0]], dt.Vertices[t.V[1]], dt.Vertices[t.V[2]]}
}

type DelaunayTriangulationOptions struct {
	Eps float64
}

type DelaunayTriangulationOption func(*DelaunayTriangulationOptions)

func WithEps(eps float64) DelaunayTriangulationOption {
	if eps <= 0 {
		panic("WithEps: eps must be non-negative")
	}

	return func(o *DelaunayTriangulationOptions) {
		o.Eps = eps
	}
}

// NOTE: All vertices must lie on a sphere.
func ComputeDelaunayTriangulation(vertices s2.PointVector, setters ...DelaunayTriangulationOption) (*DelaunayTriangulation, error) {
	opts := DelaunayTriangulationOptions{
		Eps: defaultEps,
	}
	for _, set := range setters {
		set(&opts)
	}

	numVertices := len(vertices)
	if numVertices < 4 {
		return nil,
			errors.New("s2delaunay: insufficient vertices for triangulation (minimum 4 required)")
	}
	numTriangles := 2 * (numVertices - 2)
	dt := &DelaunayTriangulation{
		Vertices:                vertices,
		Triangles:               make([]Triangle, numTriangles),
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
		return nil, errors.New("s2delaunay: inconsistent number of indices returned from QuickHull")
	}

	for _, idx := range ch.Indices {
		dt.IncidentTriangleOffsets[idx+1]++
	}
	for i := range numVertices {
		dt.IncidentTriangleOffsets[i+1] += dt.IncidentTriangleOffsets[i]
	}

	nxt := make([]int, numVertices)
	copy(nxt, dt.IncidentTriangleOffsets[:numVertices])
	for i := range numTriangles {
		base := i * 3
		for j := range 3 {
			v := ch.Indices[base+j]
			dt.Triangles[i].V[j] = v
			dt.IncidentTriangleIndices[nxt[v]] = i
			nxt[v]++
		}
		sortTriangleVerticesCCW(&dt.Triangles[i], dt.Vertices)
	}

	for i := range numVertices {
		incidentTriangles := dt.IncidentTriangles(i)
		sortIncidentTriangleIndicesCCW(i, incidentTriangles, dt.Triangles)
	}

	return dt, nil
}

func sortTriangleVerticesCCW(t *Triangle, v s2.PointVector) {
	p0, p1, p2 := v[t.V[0]], v[t.V[1]], v[t.V[2]]
	norm := p1.Sub(p0.Vector).Cross(p2.Sub(p0.Vector))
	if norm.Dot(p0.Vector) < 0 {
		t.V[1], t.V[2] = t.V[2], t.V[1]
	}
}

func sortIncidentTriangleIndicesCCW(vIdx int, incidentTris []int, tris []Triangle) {
	n := len(incidentTris)
	for i := 1; i < n; i++ {
		nxt := tris[incidentTris[i-1]].NextVertex(vIdx)
		for j := i + 1; j < n; j++ {
			prv := tris[incidentTris[j]].PrevVertex(vIdx)
			if nxt == prv {
				incidentTris[i], incidentTris[j] = incidentTris[j], incidentTris[i]
				break
			}
		}
	}
}
