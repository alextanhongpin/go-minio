package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"
)

type Image struct {
	ID        uuid.UUID
	Bucket    string
	Key       string
	Width     int64
	Height    int64
	Version   string
	Meta      json.RawMessage
	Tags      pgtype.TextArray
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ImageStore struct {
	db *sqlx.DB
}

func NewImageStore(db *sqlx.DB) *ImageStore {
	return &ImageStore{db: db}
}

type CreateRequest struct {
	Bucket  string
	Key     string
	Width   int
	Height  int
	Version string
	Tags    pgtype.TextArray
}

func (s *ImageStore) Create(ctx context.Context, req CreateRequest) (Image, error) {
	var img Image
	namedStmt, err := s.db.PrepareNamedContext(ctx, `
		INSERT INTO images(bucket, key, width, height, version, tags)
		VALUES (:bucket, :key, :width, :height, :version, :tags)
		ON CONFLICT (bucket, key) DO UPDATE SET version = EXCLUDED.version
		RETURNING *
	`)
	if err != nil {
		return img, err
	}

	err = namedStmt.Get(&img, &req)
	if err != nil {
		return img, err
	}

	return img, nil
}

type Pagination struct {
	Limit  int64
	Offset int64
}

func NewPagination() Pagination {
	return Pagination{
		Limit:  20,
		Offset: 0,
	}
}

func (s *ImageStore) FindAll(ctx context.Context, pagination Pagination) ([]Image, error) {
	var images []Image
	if err := s.db.SelectContext(ctx, &images, "SELECT * FROM images LIMIT $1 OFFSET $2", pagination.Limit, pagination.Offset); err != nil {
		return nil, err
	}
	return images, nil
}
