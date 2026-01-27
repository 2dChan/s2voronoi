package s2voronoi

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// Cell

func TestCell_SiteIndex(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c, err := vd.Cell(i)
		if err != nil {
			t.Fatalf("vd.Cell(%d) error = %v, want nil", i, err)
		}
		if got := c.SiteIndex(); got != i {
			t.Errorf("c.SiteIndex() = %v, want %v", got, i)
		}
	}
}

func TestCell_Site(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i, want := range vd.Sites {
		c, err := vd.Cell(i)
		if err != nil {
			t.Fatalf("vd.Cell(%d) error = %v, want nil", i, err)
		}
		if got := c.Site(); got != want {
			t.Errorf("c.Site() = %v, want %v", got, want)
		}
	}
}

func TestCell_NumVertices(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c, err := vd.Cell(i)
		if err != nil {
			t.Fatalf("vd.Cell(%d) error = %v, want nil", i, err)
		}
		want := vd.CellOffsets[i+1] - vd.CellOffsets[i]
		if got := c.NumVertices(); got != want {
			t.Errorf("c.NumVertices() = %v, want %v", got, want)
		}
	}
}

func TestCell_VertexIndices(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c, err := vd.Cell(i)
		if err != nil {
			t.Fatalf("vd.Cell(%d) error = %v, want nil", i, err)
		}
		want := vd.CellVertices[vd.CellOffsets[i]:vd.CellOffsets[i+1]]
		got := c.VertexIndices()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("c.VertexIndices() mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestCell_Vertex(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c, err := vd.Cell(i)
		if err != nil {
			t.Fatalf("vd.Cell(%d) error = %v, want nil", i, err)
		}
		indices := c.VertexIndices()
		for j, idx := range indices {
			want := vd.Vertices[idx]
			got, err := c.Vertex(j)
			if err != nil {
				t.Fatalf("c.Vertex(%d) error = %v, want nil", j, err)
			}
			if got != want {
				t.Errorf("c.Vertex(%d) = %v, want %v", j, got, want)
			}
		}

		if _, err := c.Vertex(-1); err == nil {
			t.Errorf("c.Vertex(-1) error = nil, want non-nil")
		}
		if _, err := c.Vertex(c.NumVertices()); err == nil {
			t.Errorf("c.Vertex(%d) error = nil, want non-nil", c.NumVertices())
		}
	}
}

func TestCell_NumNeighbors(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c, err := vd.Cell(i)
		if err != nil {
			t.Fatalf("vd.Cell(%d) error = %v, want nil", i, err)
		}
		want := vd.CellOffsets[i+1] - vd.CellOffsets[i]
		if got := c.NumNeighbors(); got != want {
			t.Errorf("c.NumNeighbors() = %v, want %v", got, want)
		}
	}
}

func TestCell_NeighborIndices(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c, err := vd.Cell(i)
		if err != nil {
			t.Fatalf("vd.Cell(%d) error = %v, want nil", i, err)
		}
		want := vd.CellNeighbors[vd.CellOffsets[i]:vd.CellOffsets[i+1]]
		got := c.NeighborIndices()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("c.NeighborIndices() mismatch (-want +got, cell %d):\n%s", i, diff)
		}
	}
}

func TestCell_Neighbor(t *testing.T) {
	vd := mustNewDiagram(t, 100)
	for i := range vd.Sites {
		c, err := vd.Cell(i)
		if err != nil {
			t.Fatalf("vd.Cell(%d) error = %v, want nil", i, err)
		}
		neighbors := c.NeighborIndices()
		for j, nIdx := range neighbors {
			got, err := c.Neighbor(j)
			if err != nil {
				t.Fatal(err)
			}
			if got.SiteIndex() != nIdx {
				t.Errorf("c.Neighbor(%d).SiteIndex() = %v, want %v", j, got.SiteIndex(), nIdx)
			}
		}
		if _, err := c.Neighbor(-1); err == nil {
			t.Errorf("c.Neighbor(-1) error = nil, want non-nil")
		}
		if _, err = c.Neighbor(c.NumNeighbors()); err == nil {
			t.Errorf("c.Neighbor(%d) error = nil, want non-nil", c.NumNeighbors())
		}
	}
}
