package util

import (
	"github.com/twpayne/go-geom"
	"github.com/uber/h3-go/v4"
	"log/slog"
)

func MultipolygonToH3(multipolygon *geom.MultiPolygon) (h3.GeoPolygon, error) {
	log := slog.Default()

	h3Loop := make([]h3.LatLng, 0)
	h3HolesLoop := make([]h3.GeoLoop, 0)
	mainpolygon := multipolygon.Polygon(0)

	log.Info("Main polygon info", slog.Int("num_rings", mainpolygon.NumLinearRings()))
	for j := 0; j < mainpolygon.NumLinearRings(); j++ {
		linearRing := mainpolygon.LinearRing(j)
		log.Info("Main polygon ring info", slog.Int("ring_index", j), slog.Int("num_coords", linearRing.NumCoords()))
		for n := 0; n < linearRing.NumCoords(); n++ {
			crd := linearRing.Coord(n)
			h3Loop = append(h3Loop, h3.NewLatLng(crd.Y(), crd.X()))
		}
	}

	for i := 1; i < multipolygon.NumPolygons(); i++ {
		polygon := multipolygon.Polygon(i)
		log.Info("Polygon info", slog.Int("polygon_index", i), slog.Int("num_rings", polygon.NumLinearRings()))
		for j := 0; j < polygon.NumLinearRings(); j++ {
			linearRing := polygon.LinearRing(j)
			log.Info("Polygon ring info", slog.Int("polygon_index", i), slog.Int("ring_index", j), slog.Int("num_coords", linearRing.NumCoords()))
			holeLoop := make([]h3.LatLng, 0)
			for n := 0; n < linearRing.NumCoords(); n++ {
				crd := linearRing.Coord(n)
				holeLoop = append(holeLoop, h3.NewLatLng(crd.Y(), crd.X()))
			}
			h3HolesLoop = append(h3HolesLoop, holeLoop)
		}
	}

	return h3.GeoPolygon{GeoLoop: h3Loop, Holes: h3HolesLoop}, nil
}
