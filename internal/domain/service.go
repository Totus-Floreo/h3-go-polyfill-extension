package domain

import "context"

// Service представляет собой интерфейс для работы с доменной логикой приложения.
type Service interface {
	Polyfill(ctx context.Context) error
}
