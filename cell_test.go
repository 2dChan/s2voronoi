// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

package s2voronoi

import (
	"testing"

	"github.com/golang/geo/r3"
	"github.com/golang/geo/s2"
	"github.com/google/go-cmp/cmp"
)

// Cell

func TestCell_SiteIndex(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		if got := c.SiteIndex(); got != i {
			t.Errorf("c.SiteIndex() = %v, want %v", got, i)
		}
	}
}

func TestCell_Site(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i, want := range vd.Sites {
		c := vd.Cell(i)
		if got := c.Site(); got != want {
			t.Errorf("c.Site() = %v, want %v", got, want)
		}
	}
}

func TestCell_NumVertices(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		want := vd.CellOffsets[i+1] - vd.CellOffsets[i]
		if got := c.NumVertices(); got != want {
			t.Errorf("c.NumVertices() = %v, want %v", got, want)
		}
	}
}

func TestCell_VertexIndices(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		want := vd.CellVertices[vd.CellOffsets[i]:vd.CellOffsets[i+1]]
		got := c.VertexIndices()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("c.VertexIndices() mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestCell_Vertex(t *testing.T) {
	assertPanic := func(c Cell, in int) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("c.Vertex(%d) did not panic, want panic", in)
			}
		}()
		c.Vertex(in)
	}

	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		indices := c.VertexIndices()
		for j, idx := range indices {
			want := vd.Vertices[idx]
			got := c.Vertex(j)
			if got != want {
				t.Errorf("c.Vertex(%d) = %v, want %v", j, got, want)
			}
		}
		assertPanic(c, -1)
		assertPanic(c, c.NumVertices())
	}
}

func TestCell_NumNeighbors(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		want := vd.CellOffsets[i+1] - vd.CellOffsets[i]
		if got := c.NumNeighbors(); got != want {
			t.Errorf("c.NumNeighbors() = %v, want %v", got, want)
		}
	}
}

func TestCell_NeighborIndices(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		want := vd.CellNeighbors[vd.CellOffsets[i]:vd.CellOffsets[i+1]]
		got := c.NeighborIndices()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("c.NeighborIndices() mismatch (-want +got, cell %d):\n%s", i, diff)
		}
	}
}

func TestCell_Neighbor(t *testing.T) {
	assertPanic := func(c Cell, in int) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("c.Neighbor(%d) did not panic, want panic", in)
			}
		}()
		c.Neighbor(in)
	}

	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c := vd.Cell(i)
		neighbors := c.NeighborIndices()
		for j, nIdx := range neighbors {
			got := c.Neighbor(j)
			if got.SiteIndex() != nIdx {
				t.Errorf("c.Neighbor(%d).SiteIndex() = %v, want %v", j, got.SiteIndex(), nIdx)
			}
		}
		assertPanic(c, -1)
		assertPanic(c, c.NumNeighbors())
	}
}

func TestCell_centroid(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.NumCells() {
		c := vd.Cell(i)
		centroid := c.centroid()

		sum := r3.Vector{X: 0, Y: 0, Z: 0}
		for j := range c.NumVertices() {
			sum = sum.Add(c.Vertex(j).Vector)
		}
		avg := sum.Mul(1.0 / float64(c.NumVertices()))
		expected := s2.Point{Vector: avg}

		if centroid.Distance(expected) > defaultEps {
			t.Errorf("c.centroid() = %v, want %v", centroid, expected)
		}
	}
}

func TestCell_centroid_Panic(t *testing.T) {
	d := &Diagram{
		Sites:         []s2.Point{s2.PointFromCoords(1, 0, 0)},
		Vertices:      []s2.Point{},
		CellNeighbors: []int{},
		CellVertices:  []int{},
		CellOffsets:   []int{0, 0},
		eps:           1e-10,
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("c.centroid() did not panic, want panic")
		}
	}()

	c := Cell{idx: 0, d: d}
	c.centroid()
}
