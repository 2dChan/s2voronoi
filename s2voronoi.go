// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

// Package s2voronoi implements Voronoi diagrams on the S2 sphere, built on Delaunay triangulation.

package s2voronoi

import (
	"errors"
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

	// eps is the numerical precision epsilon used in Voronoi diagram computations.
	eps float64
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
	if len(sites) < 4 {
		return nil, errors.New("NewDiagram: insufficient sites for diagram, minimum 4 required")
	}

	opts := &DiagramOptions{
		Eps: defaultEps,
	}
	for _, set := range setters {
		err := set(opts)
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

		eps: opts.Eps,
	}

	for i := range numTriangles {
		a, b, c := dt.TriangleVertices(i)
		d.Vertices[i] = s2.Point{Vector: triangleCircumcenter(a, b, c).Normalize()}
	}

	for vIdx := range dt.Vertices {
		offset := dt.IncidentTriangleOffsets[vIdx]
		it := dt.IncidentTriangles(vIdx)
		for i, tIdx := range it {
			nxt := s2delaunay.NextVertex(dt.Triangles[tIdx], vIdx)
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
// It panics if the index is out of range.
func (d *Diagram) Cell(i int) Cell {
	if i < 0 || i >= len(d.Sites) {
		panic(fmt.Sprintf("Cell: index %d out of range [0, %d)", i, len(d.Sites)))
	}

	return Cell{idx: i, d: d}
}

// Relax performs Lloyd's relaxation by moving sites to centroids and recomputing the diagram.
// NOTE: Allocates excessive memory by creating new Diagram per step
func (d *Diagram) Relax(steps int) error {
	if steps < 0 {
		return fmt.Errorf("Relax: steps must be non-negative, got %d", steps)
	}

	for range steps {
		for i := range d.NumCells() {
			cell := d.Cell(i)
			d.Sites[i] = s2.Point{Vector: cell.centroid().Normalize()}
		}

		// TODO: Optimize for reuse memory
		nd, err := NewDiagram(d.Sites, WithEps(d.eps))
		if err != nil {
			return err
		}

		*d = *nd
	}

	return nil
}

// triangleCircumcenter computes the circumcenter of a triangle on the sphere.
func triangleCircumcenter(a, b, c s2.Point) s2.Point {
	v1 := a.Sub(b.Vector)
	v2 := b.Sub(c.Vector)

	circumcenter := v1.Cross(v2)

	if circumcenter.Dot(a.Vector.Add(b.Vector).Add(c.Vector)) < 0 {
		circumcenter = circumcenter.Mul(-1)
	}

	return s2.Point{Vector: circumcenter}
}
