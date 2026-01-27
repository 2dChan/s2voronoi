// Copyright (c) 2026 Andrey Kriulin
// Licensed under the MIT License.
// See the LICENSE file in the project root for full license text.

// Package utils provides utility functions for generating and manipulating S2 points for Voronoi diagrams.

package utils

import (
	"math"
	"math/rand"

	"github.com/golang/geo/s1"
	"github.com/golang/geo/s2"
)

// GenerateRandomPoints generates a vector of random points on the S2 sphere.
// The seed parameter ensures reproducibility.
func GenerateRandomPoints(cnt int, seed int64) s2.PointVector {
	//nolint:gosec
	random := rand.New(rand.NewSource(seed))
	sites := make(s2.PointVector, cnt)

	for i := range cnt {
		sites[i] = s2.PointFromLatLng(s2.LatLng{
			Lat: s1.Angle((random.Float64() - 0.5) * math.Pi),
			Lng: s1.Angle((random.Float64()*2 - 1) * math.Pi),
		})
	}

	return sites
}
