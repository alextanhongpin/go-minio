package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"path/filepath"

	// NOTE: This is required for image.DecodeConfig to work
	// on different image types. This behaviour is
	// documented in the DecodeConfig example at
	// https://pkg.go.dev/image
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/kelseyhightower/envconfig"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type UploaderConfig struct {
	Endpoint        string `envconfig:"ENDPOINT" required:"true"`
	AccessKeyID     string `envconfig:"ACCESS_KEY_ID" required:"true"`
	SecretAccessKey string `envconfig:"SECRET_ACCESS_KEY" required:"true"`
	Region          string `envconfig:"REGION" required:"true"`
	Bucket          string `envconfig:"BUCKET" required:"true"`
}

func NewUploaderConfig() UploaderConfig {
	var cfg UploaderConfig
	if err := envconfig.Process("minio", &cfg); err != nil {
		log.Fatalln(err)
	}
	return cfg
}

type Uploader struct {
	client *minio.Client
	bucket string
}

func NewUploader(cfg UploaderConfig) *Uploader {
	var (
		endpoint        = cfg.Endpoint
		accessKeyID     = cfg.AccessKeyID
		secretAccessKey = cfg.SecretAccessKey
		region          = cfg.Region
		bucket          = cfg.Bucket
	)

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Region: region,
	})

	if err != nil {
		log.Fatalln(err)
	}

	return &Uploader{
		client: client,
		bucket: bucket,
	}
}

type UploadRequest struct {
	Reader   io.Reader
	Filename string
}

type UploadResponse struct {
	Bucket    string
	Key       string
	VersionID string
	Width     int
	Height    int
	Extension string
}

func (u *Uploader) Upload(ctx context.Context, req UploadRequest) (*UploadResponse, error) {
	var (
		reader = req.Reader
		bucket = u.bucket
	)
	filename, err := NewFilename(req.Filename)
	if err != nil {
		return nil, err
	}
	ext := filename.Extension()
	contentType := filename.ContentType()

	// Since image.DecodeConfig consumes the reader, we need to duplicate the
	// filestream.
	var buf bytes.Buffer
	tee, err := io.ReadAll(io.TeeReader(reader, &buf))
	if err != nil {
		return nil, err
	}

	// Skip if .svg, because this only works for jpeg, gif, and png.
	im, format, err := image.DecodeConfig(bytes.NewBuffer(tee))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to decode image", err)
	}
	if inferredExtension := "." + format; inferredExtension != ext {
		return nil, errors.New("inconsistent image extension")
	}

	// E.g. path/to/hello.png -> path/to/hello/320w.png
	newKey := func() string {
		name := fmt.Sprintf("%dw%s", im.Width, ext)
		return filepath.Join(filename.Name(), name)
	}

	key := newKey()
	uploadInfo, err := u.client.PutObject(
		ctx,
		bucket,
		key,
		&buf,
		-1,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to upload image", err)
	}
	return &UploadResponse{
		Bucket:    bucket,
		Key:       filename.Name(),
		Extension: filename.Extension(),
		Width:     im.Width,
		Height:    im.Height,
		VersionID: uploadInfo.VersionID,
	}, nil
}
