// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

package main

import (
	"fmt"
	"os"

	"github.com/2dChan/s2voronoi"
	"github.com/2dChan/s2voronoi/utils"
	svg "github.com/ajstarks/svgo"
	"github.com/golang/geo/s2"
)

const (
	filename = "voronoi.svg"

	// PlateCarreeProjection
	width  = 1500
	height = width / 2

	style     = "fill:rgb(255,255,255);stroke:rgb(170,170,170);stroke-width:1;stroke-opacity:1.0"
	siteStyle = "fill:rgb(255,0,0)"
)

func Abs(a int) int {
	if a > 0 {
		return a
	}
	return -a
}

func PointToScreen(p s2.Point) (int, int) {
	xScale := float64(width)
	proj := s2.NewPlateCarreeProjection(xScale)

	r2p := proj.Project(p)

	x := (r2p.X + xScale) / (2 * xScale)
	y := (-r2p.Y + xScale/2) / xScale

	return int(x * width), int(y * height)
}

func renderDiagram(vd *s2voronoi.Diagram) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}()

	canvas := svg.New(file)
	canvas.Start(width, height)
	canvas.Rect(0, 0, width, height, "fill:rgb(255,255,255)")

	for i := range vd.NumCells() {
		cell := vd.Cell(i)

		numPoints := cell.NumVertices()
		xPoints := make([]int, numPoints)
		yPoints := make([]int, numPoints)

		draw := true
		x0, _ := PointToScreen(cell.Vertex(0))
		for i := range cell.NumVertices() {
			xPoints[i], yPoints[i] = PointToScreen(cell.Vertex(i))
			if Abs(x0-xPoints[i]) > width/2 {
				draw = false
				break
			}
		}

		// Skip drawing boundary polygons
		if draw {
			canvas.Polygon(xPoints, yPoints, style)
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
	points := utils.GenerateRandomPoints(1000, 0)
	vd, err := s2voronoi.NewDiagram(points)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	renderDiagram(vd)
}
