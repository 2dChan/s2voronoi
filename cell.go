package s2voronoi

import (
	"fmt"

	"github.com/golang/geo/s2"
)

type Cell struct {
	idx int
	d   *Diagram
}

func (c Cell) SiteIndex() int {
	return c.idx
}

func (c Cell) Site() s2.Point {
	return c.d.Sites[c.idx]
}

func (c Cell) NumVertices() int {
	return c.d.CellOffsets[c.idx+1] - c.d.CellOffsets[c.idx]
}

func (c Cell) VertexIndices() []int {
	return c.d.CellVertices[c.d.CellOffsets[c.idx]:c.d.CellOffsets[c.idx+1]]
}

func (c Cell) Vertex(i int) (s2.Point, error) {
	start := c.d.CellOffsets[c.idx]
	end := c.d.CellOffsets[c.idx+1]
	if i < 0 || i >= end-start {
		return s2.Point{}, fmt.Errorf("Vertex: index %d out of range [0 %d)", i, end-start)
	}
	return c.d.Vertices[c.d.CellVertices[start+i]], nil
}

func (c Cell) NumNeighbors() int {
	return c.d.CellOffsets[c.idx+1] - c.d.CellOffsets[c.idx]
}

func (c Cell) NeighborIndices() []int {
	return c.d.CellNeighbors[c.d.CellOffsets[c.idx]:c.d.CellOffsets[c.idx+1]]
}

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
