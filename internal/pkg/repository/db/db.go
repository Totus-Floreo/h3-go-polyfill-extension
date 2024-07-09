package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"log"
	"os"
)

type Database interface {
}

type DatabaseImpl struct {
	log.Logger
}

func NewDatabase() Database {
	return &DatabaseImpl{}
}

func (d *DatabaseImpl) Connect() *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:qwerty@localhost:5442/taxinet")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil
	} else {
		fmt.Println("Connected to database")
	}

	return conn
}

func InsertCell(conn *pgx.Conn, cell string, resolution int, wktCell string) {
	_, err := conn.Exec(context.Background(), "INSERT INTO surges(h3_idx, resolution, geometry) VALUES ($1, $2, st_setsrid(st_wkttosql($3), 4326))", cell, resolution, wktCell)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to insert to database: %v\n", err)
		os.Exit(1)
	}
}
