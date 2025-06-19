package repository

import (
	"context"
	"fmt"
	"github.com/Totus-Floreo/h3-go-polyfill-extension/internal/domain"
	"github.com/Totus-Floreo/h3-go-polyfill-extension/internal/env"
	"github.com/Totus-Floreo/h3-go-polyfill-extension/internal/logging"
	"github.com/Totus-Floreo/h3-go-polyfill-extension/pkg/util"
	"github.com/jackc/pgx/v5"
	"github.com/uber/h3-go/v4"
	"log/slog"
)

type PostgresRepository struct {
	conn   *pgx.Conn
	logger logging.DBlogger
	env    env.Env
}

func NewPostgresRepository(ctx context.Context, logger logging.DBlogger, env env.Env) (domain.Repository, error) {
	conn, err := pgx.Connect(ctx, env.DbConnStr())
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

	return &PostgresRepository{
		conn:   conn,
		logger: logger,
		env:    env,
	}, nil
}

// Close закрывает соединение с базой данных
func (pr *PostgresRepository) Close(ctx context.Context) error {
	if pr.conn != nil {
		if err := pr.conn.Close(ctx); err != nil {
			pr.logger.Error("Unable to close database connection", slog.Any("err", err))
			return err
		}
		pr.logger.Info("Database connection closed")
	}
	return nil
}

func (pr *PostgresRepository) InsertCells(ctx context.Context, cells []h3.Cell) error {
	for _, cell := range cells {
		wktCell := util.H3ToWKT(cell)
		boundary, err := cell.Boundary()
		if err != nil {
			pr.logger.Error("Error getting cell boundary", slog.Any("err", err))
			continue
		}

		pr.logger.Debug("Cell boundary", slog.Any("boundary", boundary))
		pr.logger.Debug("WKT Cell", slog.String("wkt", wktCell))

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
		`, pr.env.TargetTable(), pr.env.OutlineTable(), pr.env.OutlineTable())
		_, err = pr.conn.Exec(ctx, insertQuery, cell.String(), cell.Resolution(), wktCell)
		if err != nil {
			pr.logger.Error("Unable to insert to database", slog.Any("err", err))
			return err
		}
	}
	return nil
}

func (pr *PostgresRepository) GetOutline(ctx context.Context) ([]byte, error) {
	var outlineWKB []byte

	query := fmt.Sprintf("SELECT ST_AsBinary(geometry) FROM %s LIMIT 1", pr.env.OutlineTable())
	err := pr.conn.QueryRow(ctx, query).Scan(&outlineWKB)
	if err != nil {
		pr.logger.Error("Unable to get outline geometry", slog.Any("err", err))
		return nil, err
	}

	return outlineWKB, nil
}
