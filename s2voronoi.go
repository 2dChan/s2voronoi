// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

// Package s2voronoi implements Voronoi diagrams on the S2 sphere, built on Delaunay triangulation.

package s2voronoi

import (
	"fmt"

	"github.com/2dChan/s2voronoi/s2delaunay"
	"github.com/golang/geo/s2"
)

const (
	defaultEps = 1e-12
)

// Diagram represents a Voronoi diagram on the S2 sphere.
type Diagram struct {
	// Sites are the input points on the unit sphere.
	Sites s2.PointVector
	// Vertices are the Voronoi vertices on the unit sphere.
	Vertices s2.PointVector

	// CellVertices contains indices of vertices for each cell, sorted in CCW order,
	// forming a CSR-like sparse representation.
	CellVertices []int
	// CellNeighbors contains indices of neighboring sites for each cell, sorted in CCW order,
	// forming a CSR-like sparse representation.
	CellNeighbors []int
	// CellOffsets contains offsets for slicing cell data in a CSR-like format.
	CellOffsets []int
}

// DiagramOptions holds configuration options for Voronoi diagram creation.
type DiagramOptions struct {
	Eps float64
}

// DiagramOption is a functional option type for Voronoi diagram configuration.
type DiagramOption func(*DiagramOptions) error

// WithEps sets the numerical precision epsilon for Voronoi diagram computation.
// It must be positive.
func WithEps(eps float64) DiagramOption {
	return func(o *DiagramOptions) error {
		if eps <= 0 {
			return fmt.Errorf("WithEps: eps must be positive got %v", eps)

		}
		o.Eps = eps
		return nil
	}
}

// NewDiagram creates a new Voronoi diagram from the given sites.
// The sites must lie on the unit sphere, there must be at least 4 sites, and they must not be coplanar.
// It returns an error if the diagram cannot be constructed.
func NewDiagram(sites s2.PointVector, setters ...DiagramOption) (*Diagram, error) {
	opts := DiagramOptions{
		Eps: defaultEps,
	}
	for _, set := range setters {
		err := set(&opts)
		if err != nil {
			return nil, err
		}
	}

	dt, err := s2delaunay.NewTriangulation(sites, s2delaunay.WithEps(opts.Eps))
	if err != nil {
		return nil, err
	}

	numTriangles := len(dt.Triangles)
	numNeighbors := len(dt.IncidentTriangleIndices)
	d := &Diagram{
		Sites:         dt.Vertices,
		Vertices:      make(s2.PointVector, numTriangles),
		CellVertices:  dt.IncidentTriangleIndices,
		CellNeighbors: make([]int, numNeighbors),
		CellOffsets:   dt.IncidentTriangleOffsets,
	}

	for i := range numTriangles {
		p, err := dt.TriangleVertices(i)
		if err != nil {
			return nil, err
		}
		d.Vertices[i] = s2.Point{Vector: triangleCircumcenter(p[0], p[1], p[2]).Normalize()}
	}

	for vIdx := range dt.Vertices {
		offset := dt.IncidentTriangleOffsets[vIdx]
		it, err := dt.IncidentTriangles(vIdx)
		if err != nil {
			return nil, err
		}
		for i, tIdx := range it {
			nxt, err := s2delaunay.NextVertex(dt.Triangles[tIdx], vIdx)
			if err != nil {
				return nil, err
			}
			d.CellNeighbors[offset+i] = nxt
		}
	}

	return d, nil
}

// NumCells returns the number of cells in the diagram.
func (d *Diagram) NumCells() int {
	return len(d.Sites)
}

// Cell returns the Voronoi cell at the specified index.
// It returns an error if the index is out of range.
func (d *Diagram) Cell(i int) (Cell, error) {
	if i < 0 || i >= len(d.Sites) {
		return Cell{}, fmt.Errorf("Cell: index %d out of range [0, %d)", i, len(d.Sites))
	}

	return Cell{idx: i, d: d}, nil
}

// triangleCircumcenter computes the circumcenter of a triangle on the sphere.
func triangleCircumcenter(p1, p2, p3 s2.Point) s2.Point {
	v1 := p1.Sub(p2.Vector)
	v2 := p2.Sub(p3.Vector)

	circumcenter := v1.Cross(v2)

	if circumcenter.Dot(p1.Vector.Add(p2.Vector).Add(p3.Vector)) < 0 {
		circumcenter = circumcenter.Mul(-1)
	}

	return s2.Point{Vector: circumcenter}
}
