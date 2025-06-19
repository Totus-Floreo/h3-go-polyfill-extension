package app

import (
	"context"
	"fmt"
	"github.com/Totus-Floreo/h3-go-polyfill-extension/internal/domain"
	"github.com/Totus-Floreo/h3-go-polyfill-extension/internal/logging"
	"github.com/Totus-Floreo/h3-go-polyfill-extension/pkg/util"
	"log/slog"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkb"
	"github.com/uber/h3-go/v4"
)

type Service struct {
	logger     logging.Logger
	repository domain.Repository
}

func NewService(logger logging.Logger, repository domain.Repository) domain.Service {
	return &Service{
		logger:     logger,
		repository: repository,
	}
}

func (s *Service) getMultipolygon(ctx context.Context) (*geom.MultiPolygon, error) {
	outlineWKB, err := s.repository.GetOutline(ctx)
	if err != nil {
		s.logger.Error("Unable to get outline WKB", slog.Any("err", err))
		return nil, err
	}

	geometry, err := wkb.Unmarshal(outlineWKB)
	if err != nil {
		s.logger.Error("Unable to unmarshal WKB", slog.Any("err", err))
		return nil, err
	}

	multipolygon, ok := geometry.(*geom.MultiPolygon)
	if !ok {
		s.logger.Error("Geometry is not a multipolygon")
		return nil, fmt.Errorf("not a multipolygon")
	}

	s.logger.Info("Multipolygon loaded", slog.Int("num_polygons", multipolygon.NumPolygons()))
	return multipolygon, nil
}

// Polyfill теперь принимает параметры
func (s *Service) Polyfill(ctx context.Context) error {
	multipolygon, err := s.getMultipolygon(ctx)
	if err != nil {
		return fmt.Errorf("failed to get multipolygon: %w", err)
	}

	h3polygon, err := util.MultipolygonToH3(multipolygon)
	if err != nil {
		s.logger.Error("Failed to convert multipolygon to H3", slog.Any("err", err))
		return fmt.Errorf("failed to convert multipolygon to H3: %w", err)
	}

	cells, err := h3.PolygonToCellsExperimental(h3polygon, 7, h3.ContainmentOverlappingBbox)
	if err != nil {
		s.logger.Error("Error during polygon to cells conversion", slog.Any("err", err))
		return fmt.Errorf("error during polygon to cells conversion: %w", err)
	}
	s.logger.Info("Polygon to cells conversion done", slog.Int("num_cells", len(cells)))

	if err := s.repository.InsertCells(ctx, cells); err != nil {
		s.logger.Error("Error inserting cells", slog.Any("err", err))
		return fmt.Errorf("error inserting cells: %w", err)
	}

	s.logger.Info("Polyfill done")
	return nil
}
