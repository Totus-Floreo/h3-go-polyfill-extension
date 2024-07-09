package app

import (
	"context"
	"fmt"
	"github.com/Totus-Floreo/h3-go"
	"github.com/jackc/pgx/v5"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
	"os"
)

func Polyfill() {
	var geometry geom.T
	if err := geojson.Unmarshal([]byte(GeoJSON), &geometry); err != nil {
		fmt.Println(err)
		return
	}

	multipolygon, ok := geometry.(*geom.MultiPolygon)
	if !ok {
		fmt.Println("not a multipolygon")
		return
	}

	fmt.Println("multipolygon have a ", multipolygon.NumPolygons(), " polygons")

	h3Loop := make([]h3.LatLng, 0)
	h3HolesLoop := make([]h3.GeoLoop, 0)

	mainpolygon := multipolygon.Polygon(0)
	fmt.Println("main polygon have a ", mainpolygon.NumLinearRings(), " rings")
	for j := 0; j < mainpolygon.NumLinearRings(); j++ {
		linearRing := mainpolygon.LinearRing(j)
		fmt.Println("linearRing ", j, " have a ", linearRing.NumCoords(), " coords")
		for n := 0; n < linearRing.NumCoords(); n++ {
			crd := linearRing.Coord(n)
			h3Loop = append(h3Loop, h3.NewLatLng(crd.Y(), crd.X()))
		}
	}

	for i := 1; i < multipolygon.NumPolygons(); i++ {
		polygon := multipolygon.Polygon(i)
		fmt.Println("polygon ", i, " have a ", polygon.NumLinearRings(), " rings")
		for j := 0; j < polygon.NumLinearRings(); j++ {
			linearRing := polygon.LinearRing(j)
			fmt.Println("linearRing ", j, " have a ", linearRing.NumCoords(), " coords")
			h3Loop := make([]h3.LatLng, 0)
			for n := 0; n < linearRing.NumCoords(); n++ {
				crd := linearRing.Coord(n)
				h3Loop = append(h3Loop, h3.NewLatLng(crd.Y(), crd.X()))
			}
			h3HolesLoop = append(h3HolesLoop, h3Loop)
		}
	}

	ctx := context.Background()
	fmt.Println(IsUserAuthorized(ctx))
	ctx := context.WithValue(ctx, "TEST", "admin")

	h3polygon := h3.GeoPolygon{GeoLoop: h3Loop, Holes: h3HolesLoop}

	cells := h3.PolygonToCellsExperimental(h3polygon, h3.PolyfillModeOverlapping, 7)
	fmt.Println("Amount of cells : ", len(cells))

	conn, err := pgx.Connect(context.Background(), "postgres://postgres:qwerty@localhost:5442/taxinet")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	if err := conn.Ping(context.Background()); err != nil {
		return
	} else {
		fmt.Println("Connected to database")
	}

	fmt.Println("Amount of cells : ", len(cells))

	for _, cell := range cells {

		wktCell := h3ToWKT(cell)

		fmt.Println(cell.Boundary())
		fmt.Println("WKT Cell :", wktCell)

		_, err = conn.Exec(context.Background(), "INSERT INTO surges(h3_idx, resolution, geometry) VALUES ($1, $2, st_setsrid(st_wkttosql($3), 4326))", cell.String(), cell.Resolution(), wktCell)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to insert to database: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Polyfill done")
}

func h3ToWKT(cell h3.Cell) string {
	boundary := cell.Boundary()

	// Начинаем с "POLYGON(("
	wkt := "POLYGON(("

	// Добавляем все координаты
	for i, crd := range boundary {
		// Добавляем координаты в формате "долгота широта"
		wkt += fmt.Sprintf("%f %f", crd.Lng, crd.Lat)

		// Если это не последняя координата, добавляем запятую
		if i != len(boundary)-1 {
			wkt += ", "
		}
	}

	// Добавляем первую координату в конец, чтобы закрыть кольцо
	wkt += fmt.Sprintf(", %f %f", boundary[0].Lng, boundary[0].Lat)

	// Закрываем скобки
	wkt += "))"

	return wkt
}
