package domain

import (
	"context"
	"github.com/uber/h3-go/v4"
)

// Repository представляет собой интерфейс для работы с репозиториями в доменной логике приложения.
type Repository interface {
	InsertCells(ctx context.Context, cells []h3.Cell) error
	GetOutline(ctx context.Context) ([]byte, error)
	Close(ctx context.Context) error
}
