// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

package main

import (
	"log"
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

	for i := range vd.NumCells() {
		cell, err := vd.Cell(i)
		if err != nil {
			log.Printf("error getting cell %d: %v", i, err)
			continue
		}

		numPoints := cell.NumVertices()
		xPoints := make([]int, numPoints)
		yPoints := make([]int, numPoints)

		draw := true
		v, err := cell.Vertex(0)
		if err != nil {
			log.Printf("error getting vertex for cell %d: %v", i, err)
			continue
		}
		x0, _ := PointToScreen(v)
		for j := range cell.NumVertices() {
			v, err := cell.Vertex(j)
			if err != nil {
				log.Printf("error getting vertex for cell %d: %v", i, err)
				draw = false
				break
			}

			xPoints[j], yPoints[j] = PointToScreen(v)
			if Abs(x0-xPoints[j]) > width/2 {
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
		cell, err := vd.Cell(i)
		if err != nil {
			log.Printf("error getting cell %d: %v", i, err)
			continue
		}
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
		log.Fatal(err)
	}

	renderDiagram(vd)
}
