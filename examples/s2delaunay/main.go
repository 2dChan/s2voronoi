// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

package main

import (
	"log"
	"math"
	"os"

	"github.com/2dChan/s2voronoi/s2delaunay"
	"github.com/2dChan/s2voronoi/utils"
	svg "github.com/ajstarks/svgo"
	"github.com/golang/geo/s2"
)

const (
	filename = "delaunay.svg"

	// PlateCarreeProjection
	width  = 1500
	height = width / 2

	polygonStyle = "fill:rgb(255,255,255);stroke:rgb(170,170,170);stroke-width:1;stroke-opacity:1.0"
	siteStyle    = "fill:rgb(0,0,255)"
)

func PointToScreen(p s2.Point) (int, int) {
	xScale := float64(width)
	proj := s2.NewPlateCarreeProjection(xScale)

	r2p := proj.Project(p)

	x := (r2p.X + xScale) / (2 * xScale)
	y := (-r2p.Y + xScale/2) / xScale

	return int(x * width), int(y * height)
}

func renderTriangulation(dt *s2delaunay.Triangulation) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	canvas := svg.New(file)
	canvas.Start(width, height)
	canvas.Rect(0, 0, width, height, "fill:rgb(255,255,255)")

	xPoints := make([]int, 0)
	yPoints := make([]int, 0)
	for _, tri := range dt.Triangles {
		xPoints = xPoints[:0]
		yPoints = yPoints[:0]

		draw := true
		v0 := dt.Vertices[tri[0]]
		lng0 := s2.LatLngFromPoint(v0).Lng.Radians()
		for _, id := range tri {
			v := dt.Vertices[id]
			lng := s2.LatLngFromPoint(v).Lng.Radians()
			if math.Abs(lng0-lng) > math.Pi {
				draw = false
				break
			}

			x, y := PointToScreen(v)
			xPoints = append(xPoints, x)
			yPoints = append(yPoints, y)
		}

		// Skip polygons that may cross the antimeridian to avoid rendering issues
		if draw {
			canvas.Polygon(xPoints, yPoints, polygonStyle)
		}
	}

	for _, p := range dt.Vertices {
		x, y := PointToScreen(p)
		canvas.Circle(x, y, 3, siteStyle)
	}
	canvas.End()
}

func main() {
	const (
		numPoints = 1000
		seed      = 0
	)

	points := utils.GenerateRandomPoints(numPoints, seed)
	dt, err := s2delaunay.NewTriangulation(points)
	if err != nil {
		log.Fatal(err)
	}

	renderTriangulation(dt)
}
