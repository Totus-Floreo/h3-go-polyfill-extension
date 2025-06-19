package app

import (
	"context"
	"fmt"
	"github.com/Totus-Floreo/h3-go-polyfill-extension/pkg/util"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkb"
	"github.com/uber/h3-go/v4"
)

func connectDB(ctx context.Context, dbConnStr string, logger *slog.Logger) (*pgx.Conn, error) {
	conn, err := pgx.Connect(ctx, dbConnStr)
	if err != nil {
		logger.Error("Unable to connect to database", slog.Any("err", err))
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		logger.Error("Unable to ping database", slog.Any("err", err))
		conn.Close(ctx)
		return nil, err
	}

	logger.Info("Connected to database")

	return conn, nil
}

func getMultipolygon(ctx context.Context, conn *pgx.Conn, outlineTable string, logger *slog.Logger) (*geom.MultiPolygon, error) {
	var outlineWKB []byte

	query := fmt.Sprintf("SELECT ST_AsBinary(geometry) FROM %s LIMIT 1", outlineTable)
	err := conn.QueryRow(ctx, query).Scan(&outlineWKB)
	if err != nil {
		logger.Error("Unable to get outline geometry", slog.Any("err", err))
		return nil, err
	}

	geometry, err := wkb.Unmarshal(outlineWKB)
	if err != nil {
		logger.Error("Unable to unmarshal WKB", slog.Any("err", err))
		return nil, err
	}

	multipolygon, ok := geometry.(*geom.MultiPolygon)
	if !ok {
		logger.Error("Geometry is not a multipolygon")
		return nil, fmt.Errorf("not a multipolygon")
	}

	logger.Info("Multipolygon loaded", slog.Int("num_polygons", multipolygon.NumPolygons()))
	return multipolygon, nil
}

func multipolygonToH3(multipolygon *geom.MultiPolygon, logger *slog.Logger) (h3.GeoPolygon, error) {
	h3Loop := make([]h3.LatLng, 0)
	h3HolesLoop := make([]h3.GeoLoop, 0)
	mainpolygon := multipolygon.Polygon(0)

	logger.Info("Main polygon info", slog.Int("num_rings", mainpolygon.NumLinearRings()))
	for j := 0; j < mainpolygon.NumLinearRings(); j++ {
		linearRing := mainpolygon.LinearRing(j)
		logger.Info("Main polygon ring info", slog.Int("ring_index", j), slog.Int("num_coords", linearRing.NumCoords()))
		for n := 0; n < linearRing.NumCoords(); n++ {
			crd := linearRing.Coord(n)
			h3Loop = append(h3Loop, h3.NewLatLng(crd.Y(), crd.X()))
		}
	}

	for i := 1; i < multipolygon.NumPolygons(); i++ {
		polygon := multipolygon.Polygon(i)
		logger.Info("Polygon info", slog.Int("polygon_index", i), slog.Int("num_rings", polygon.NumLinearRings()))
		for j := 0; j < polygon.NumLinearRings(); j++ {
			linearRing := polygon.LinearRing(j)
			logger.Info("Polygon ring info", slog.Int("polygon_index", i), slog.Int("ring_index", j), slog.Int("num_coords", linearRing.NumCoords()))
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

func insertCells(ctx context.Context, conn *pgx.Conn, cells []h3.Cell, targetTable, outlineTable string, logger *slog.Logger) error {
	for _, cell := range cells {
		wktCell := util.H3ToWKT(cell)
		boundary, err := cell.Boundary()
		if err != nil {
			logger.Error("Error getting cell boundary", slog.Any("err", err))
			continue
		}

		logger.Debug("Cell boundary", slog.Any("boundary", boundary))
		logger.Debug("WKT Cell", slog.String("wkt", wktCell))

		insertQuery := fmt.Sprintf(`
			INSERT INTO %s(h3_idx, resolution, geometry)
			SELECT $1, $2, ST_Intersection(
				st_setsrid(st_wkttosql($3), 4326),
				(SELECT geometry FROM %s LIMIT 1)
			)
			WHERE ST_Intersects(
				st_setsrid(st_wkttosql($3), 4326),
				(SELECT geometry FROM %s LIMIT 1)
			)
		`, targetTable, outlineTable, outlineTable)
		_, err = conn.Exec(ctx, insertQuery, cell.String(), cell.Resolution(), wktCell)
		if err != nil {
			logger.Error("Unable to insert to database", slog.Any("err", err))
			return err
		}
	}
	return nil
}

// Polyfill теперь принимает параметры
func Polyfill(targetTable, outlineTable, dbConnStr string) {
	ctx := context.Background()
	logger := slog.Default()

	conn, err := connectDB(ctx, dbConnStr, logger)
	if err != nil {
		return
	}
	defer conn.Close(ctx)

	multipolygon, err := getMultipolygon(ctx, conn, outlineTable, logger)
	if err != nil {
		return
	}

	h3polygon, err := multipolygonToH3(multipolygon, logger)
	if err != nil {
		logger.Error("Failed to convert multipolygon to H3", slog.Any("err", err))
		return
	}

	cells, err := h3.PolygonToCellsExperimental(h3polygon, 7, h3.ContainmentOverlappingBbox)
	if err != nil {
		logger.Error("Error during polygon to cells conversion", slog.Any("err", err))
		return
	}
	logger.Info("Polygon to cells conversion done", slog.Int("num_cells", len(cells)))

	if err := insertCells(ctx, conn, cells, targetTable, outlineTable, logger); err != nil {
		logger.Error("Error inserting cells", slog.Any("err", err))
		return
	}

	logger.Info("Polyfill done")
}
