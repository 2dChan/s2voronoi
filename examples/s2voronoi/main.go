// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

package main

import (
	"log"
	"math"
	"os"

	"github.com/2dChan/s2voronoi"
	"github.com/2dChan/s2voronoi/utils"
	svg "github.com/ajstarks/svgo"
	"github.com/golang/geo/s2"
)

const (
	filename = "voronoi.svg"

	width  = 1500
	height = width / 2

	polygonStyle = "fill:rgb(255,255,255);stroke:rgb(170,170,170);stroke-width:1;stroke-opacity:1.0"
	siteStyle    = "fill:rgb(255,0,0)"
)

func PointToScreen(p s2.Point) (int, int) {
	xScale := float64(width)
	proj := s2.NewMercatorProjection(xScale)

	r2p := proj.Project(p)

	x := (r2p.X + xScale) / (2 * xScale)
	y := (-r2p.Y + xScale/2) / xScale

	return int(x * width), int(y * height)
}

func renderDiagram(vd *s2voronoi.Diagram) {
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
	for i := range vd.NumCells() {
		cell := vd.Cell(i)
		xPoints := xPoints[:0]
		yPoints := yPoints[:0]

		draw := true
		sLng := s2.LatLngFromPoint(cell.Site()).Lng.Radians()
		for j := range cell.NumVertices() {
			vert := cell.Vertex(j)
			vLng := s2.LatLngFromPoint(vert).Lng.Radians()
			if math.Abs(vLng-sLng) > math.Pi {
				draw = false
				break
			}

			x, y := PointToScreen(vert)
			xPoints = append(xPoints, x)
			yPoints = append(yPoints, y)
		}

		// Skip polygons that may cross the antimeridian to avoid rendering issues
		if draw {
			canvas.Polygon(xPoints, yPoints, polygonStyle)
		}
	}

	for i := range vd.NumCells() {
		cell := vd.Cell(i)
		site := cell.Site()
		sx, sy := PointToScreen(site)
		canvas.Circle(sx, sy, 3, siteStyle)
	}
	canvas.End()
}

func main() {
	const (
		numPoints  = 1000
		seed       = 0
		relaxSteps = 5
	)

	points := utils.GenerateRandomPoints(numPoints, seed)
	vd, err := s2voronoi.NewDiagram(points)
	if err != nil {
		log.Fatal(err)
	}

	err = vd.Relax(relaxSteps)
	if err != nil {
		log.Fatal(err)
	}

	renderDiagram(vd)
}
