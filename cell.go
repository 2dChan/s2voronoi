// Package s2voronoi implements Voronoi diagrams on the S2 sphere, built on Delaunay triangulation.

package s2voronoi

import (
	"fmt"

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
// It returns an error if the index is out of range.
func (c Cell) Vertex(i int) (s2.Point, error) {
	start := c.d.CellOffsets[c.idx]
	end := c.d.CellOffsets[c.idx+1]
	if i < 0 || i >= end-start {
		return s2.Point{}, fmt.Errorf("Vertex: index %d out of range [0 %d)", i, end-start)
	}
	return c.d.Vertices[c.d.CellVertices[start+i]], nil
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
// It returns an error if the index is out of range.
func (c Cell) Neighbor(i int) (Cell, error) {
	start := c.d.CellOffsets[c.idx]
	end := c.d.CellOffsets[c.idx+1]
	if i < 0 || i >= end-start {
		return Cell{}, fmt.Errorf("Neighbor: index %d out of range [0 %d)", i, end-start)
	}
	nc, err := c.d.Cell(c.d.CellNeighbors[start+i])
	if err != nil {
		return Cell{}, err
	}
	return nc, nil
}
