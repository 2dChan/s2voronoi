// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

package s2voronoi

import (
	"github.com/2dChan/s2voronoi/s2delaunay"
	"github.com/golang/geo/s2"
)

const (
	defaultEps = 1e-12
)

type VoronoiDiagram struct {
	Sites    s2.PointVector
	Vertices s2.PointVector

	// NOTE: Sort in CCW per Cell(look out of sphere)
	CellVertices []int
	// NOTE: Sort in CCW per Cell(look out of sphere)
	CellNeighbors []int
	CellOffsets   []int
}

func (vd *VoronoiDiagram) NumCells() int {
	return len(vd.Sites)
}

func (vd *VoronoiDiagram) Cell(i int) Cell {
	return Cell{idx: i, vd: vd}
}

func ComputeVoronoiDiagram(sites s2.PointVector, eps float64) (*VoronoiDiagram, error) {
	if eps == 0 {
		eps = defaultEps
	}

	dt, err := s2delaunay.ComputeDelaunayTriangulation(sites, s2delaunay.WithEps(eps))
	if err != nil {
		return nil, err
	}

	numTriangles := len(dt.Triangles)
	numNeighbors := len(dt.IncidentTriangleIndices)
	vd := &VoronoiDiagram{
		Sites:         dt.Vertices,
		Vertices:      make(s2.PointVector, numTriangles),
		CellVertices:  dt.IncidentTriangleIndices,
		CellNeighbors: make([]int, numNeighbors),
		CellOffsets:   dt.IncidentTriangleOffsets,
	}

	for i := range numTriangles {
		p := dt.TriangleVertices(i)
		vd.Vertices[i] = s2.Point{Vector: triangleCircumcenter(p[0], p[1], p[2]).Normalize()}
	}

	for vIdx := range dt.Vertices {
		offset := dt.IncidentTriangleOffsets[vIdx]
		it := dt.IncidentTriangles(vIdx)
		for i, tIdx := range it {
			vd.CellNeighbors[offset+i] = dt.Triangles[tIdx].NextVertex(vIdx)
		}
	}

	return vd, nil
}

type Cell struct {
	idx int
	vd  *VoronoiDiagram
}

func (c Cell) SiteIndex() int {
	return c.idx
}

func (c Cell) Site() s2.Point {
	return c.vd.Sites[c.idx]
}

func (c Cell) NumVertices() int {
	return c.vd.CellOffsets[c.idx+1] - c.vd.CellOffsets[c.idx]
}

func (c Cell) VertexIndices() []int {
	return c.vd.CellVertices[c.vd.CellOffsets[c.idx]:c.vd.CellOffsets[c.idx+1]]
}

func (c Cell) Vertex(i int) s2.Point {
	start := c.vd.CellOffsets[c.idx]
	end := c.vd.CellOffsets[c.idx+1]
	if i < 0 || i > end-start {
		panic("Vertex: index out of range")
	}
	return c.vd.Vertices[c.vd.CellVertices[start+i]]
}

func (c Cell) NumNeighbors() int {
	return c.vd.CellOffsets[c.idx+1] - c.vd.CellOffsets[c.idx]
}

func (c Cell) NeighborIndices() []int {
	return c.vd.CellNeighbors[c.vd.CellOffsets[c.idx]:c.vd.CellOffsets[c.idx+1]]
}

func (c Cell) Neighbor(i int) Cell {
	start := c.vd.CellOffsets[c.idx]
	end := c.vd.CellOffsets[c.idx+1]
	if i < 0 || i > end-start {
		panic("Neighbor: index out of range")
	}
	return c.vd.Cell(c.vd.CellNeighbors[start+i])
}

func triangleCircumcenter(p1, p2, p3 s2.Point) s2.Point {
	v1 := p1.Sub(p2.Vector)
	v2 := p2.Sub(p3.Vector)

	circumcenter := v1.Cross(v2)

	if circumcenter.Dot(p1.Vector.Add(p2.Vector).Add(p3.Vector)) < 0 {
		circumcenter = circumcenter.Mul(-1)
	}

	return s2.Point{Vector: circumcenter}
}
