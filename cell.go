// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

// Package s2voronoi implements Voronoi diagrams on the S2 sphere, built on Delaunay triangulation.

package s2voronoi

import (
	"fmt"

	"github.com/golang/geo/r3"
	"github.com/golang/geo/s2"
)

// Cell represents a Voronoi cell. It is a view structure for accessing a cell in a Diagram.
// The cell's index corresponds to the index of its site in the Diagram's Sites.
type Cell struct {
	idx int
	d   *Diagram
}

// SiteIndex returns the index of the site in the Diagram's Sites.
func (c Cell) SiteIndex() int {
	return c.idx
}

// Site returns the site point of the cell.
func (c Cell) Site() s2.Point {
	return c.d.Sites[c.idx]
}

// NumVertices returns the number of vertices in the cell.
// This equals the number of neighbors.
func (c Cell) NumVertices() int {
	return c.d.CellOffsets[c.idx+1] - c.d.CellOffsets[c.idx]
}

// VertexIndices returns the indices of the vertices that form the cell in the Diagram's Vertices,
// sorted in counter-clockwise order when looking out of the sphere.
func (c Cell) VertexIndices() []int {
	return c.d.CellVertices[c.d.CellOffsets[c.idx]:c.d.CellOffsets[c.idx+1]]
}

// Vertex returns the vertex at the specified index.
// It panics if the index is out of range.
func (c Cell) Vertex(i int) s2.Point {
	start := c.d.CellOffsets[c.idx]
	end := c.d.CellOffsets[c.idx+1]
	if i < 0 || i >= end-start {
		panic(fmt.Sprintf("Vertex: index %d out of range [0 %d)", i, end-start))
	}
	return c.d.Vertices[c.d.CellVertices[start+i]]
}

// NumNeighbors returns the number of neighboring cells.
// This equals the number of vertices.
func (c Cell) NumNeighbors() int {
	return c.d.CellOffsets[c.idx+1] - c.d.CellOffsets[c.idx]
}

// NeighborIndices returns the indices of the neighboring cells in the Diagram,
// sorted in counter-clockwise order when looking out of the sphere.
func (c Cell) NeighborIndices() []int {
	return c.d.CellNeighbors[c.d.CellOffsets[c.idx]:c.d.CellOffsets[c.idx+1]]
}

// Neighbor returns the neighboring cell at the specified index.
// It panics if the index is out of range.
func (c Cell) Neighbor(i int) Cell {
	start := c.d.CellOffsets[c.idx]
	end := c.d.CellOffsets[c.idx+1]
	if i < 0 || i >= end-start {
		panic(fmt.Sprintf("Neighbor: index %d out of range [0 %d)", i, end-start))
	}
	nc := c.d.Cell(c.d.CellNeighbors[start+i])
	return nc
}

// centroid returns the centroid of the cell by averaging its vertex vectors on the unit sphere.
func (c Cell) centroid() s2.Point {
	num := c.NumVertices()
	if num == 0 {
		panic("centroid: cell has no vertices")
	}

	sum := r3.Vector{X: 0, Y: 0, Z: 0}
	for i := range num {
		sum = sum.Add(c.Vertex(i).Vector)
	}
	avg := sum.Mul(1.0 / float64(num))

	return s2.Point{Vector: avg.Normalize()}
}
