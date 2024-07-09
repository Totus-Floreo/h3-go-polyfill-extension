package util

import (
	"fmt"
	"github.com/Totus-Floreo/h3-go"
)

func H3ToWKT(cell h3.Cell) string {
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
