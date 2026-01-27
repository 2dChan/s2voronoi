# s2voronoi

[![Build Status](https://github.com/2dChan/s2voronoi/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/2dChan/s2voronoi/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/2dChan/s2voronoi)](https://goreportcard.com/report/github.com/2dChan/s2voronoi)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/2dChan/s2voronoi)](https://pkg.go.dev/github.com/2dChan/s2voronoi)

A Go library for computing Voronoi diagrams on the S2 sphere using Delaunay triangulation, with seamless integration with the golang/geo S2 library (utilizing S2 types).

## Installation

To install s2voronoi, use go get:

```bash
go get github.com/2dChan/s2voronoi
```

## Quick Start

Import the library and create a Voronoi diagram from points on the sphere:

```go
package main

import (
    "log"
    "github.com/2dChan/s2voronoi"
    "github.com/2dChan/s2voronoi/utils"
    "github.com/golang/geo/s2"
)

func main() {
    // Generate random points on the sphere
    points := utils.GenerateRandomPoints(100, 0)

    // Create the Voronoi diagram
    diagram, err := s2voronoi.NewDiagram(points)
    if err != nil {
        log.Fatal(err)
    }

    // Iterate over cells
    for i := 0; i < diagram.NumCells(); i++ {
        cell := diagram.Cell(i)
        site := cell.Site()
        fmt.Printf("Cell %d site: %v, has %d vertices\n", i, site, cell.NumVertices())
    }
}
```

See examples for detailed usage:

- **Basic Diagram Generation**: [s2voronoi](examples/s2voronoi/main.go) - Generates a Voronoi diagram and exports to SVG.
- **Delaunay Triangulation**: [s2delaunay](examples/s2delaunay/main.go) - Generates a Delaunay Triangulation and exports to SVG.

Run an example:

```bash
go run examples/s2voronoi/main.go
```

## License

The code is distributed under the [MIT license](LICENSE).

## Links

- [golang/geo](https://github.com/golang/geo) - Spherical geometry library
