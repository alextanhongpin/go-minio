package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	minSize = 1024            // 1KB
	maxSize = 1024 * 1024 * 5 // 5MB
	// Overrides the default filename, so that we do not need to worry about
	// encoding them, and the URL looks prettier too.
	fileName = "default"

	// Default 1 hour expiry for POST presigned URL.
	presignedURLValidity = time.Hour
)

func main() {
	endpoint := "localhost:9000"
	accessKeyID := "minio"
	secretAccessKey := "minio123"
	useSSL := false
	region := "ap-southeast-1"

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
		Region: region,
	})

	if err != nil {
		log.Fatal(err)
	}

	bucket := "mybucket"
	if err := createBucket(client, region, bucket); err != nil {
		log.Printf("createBucketErr: %s\n", err)
	}
	if err := presignedPostPolicy(client, PresignedPostPolicyRequest{
		Bucket: "mybucket",
		Key:    "hello.png",
	}); err != nil {
		log.Printf("postPolicyErr: %v\n", err)
	}
}

func createBucket(client *minio.Client, region, bucket string) error {
	ctx := context.Background()
	found, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if found {
		fmt.Println("bucket exists")
		return nil
	}

	// Create a bucket at region 'us-east-1' with object locking enabled.
	err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: region})
	if err != nil {
		return err
	}
	fmt.Println("Successfully created mybucket.")
	return nil
}

type PresignedPostPolicyRequest struct {
	Bucket   string
	Prefix   string
	Key      string
	Duration time.Duration
}

func presignedPostPolicy(client *minio.Client, req PresignedPostPolicyRequest) error {
	if req.Duration == time.Duration(0) {
		req.Duration = presignedURLValidity
	}

	// Only allow certain image type.
	ext := filepath.Ext(req.Key)
	contentType := mime.TypeByExtension(ext)
	if !strings.HasPrefix(contentType, "image/") {
		return errors.New("Content-Type invalid")
	}

	policy := minio.NewPostPolicy()
	policy.SetBucket(req.Bucket)

	buildKey := func() string {
		// <bucket>/<prefix?>/<uuid><extension>
		return filepath.Join(req.Prefix, fmt.Sprintf("%s%s", uuid.New().String(), ext))
	}

	// Overrides all the filename to default.png
	policy.SetKey(buildKey())
	policy.SetExpires(time.Now().UTC().Add(req.Duration)) // Expires in 1 day.

	policy.SetContentType(contentType)

	// Only allow content size in range 1BK to 5MB.
	policy.SetContentLengthRange(minSize, maxSize)

	// Add a user metadata using the key "custom" and value "user".
	policy.SetUserMetadata("custom", "user")

	ctx := context.Background()
	// Get the POST form key/value object.
	url, formData, err := client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return err
	}

	fmt.Println(url)
	for k, v := range formData {
		fmt.Printf("-F %s=%s\n", k, v)
	}

	return nil
}
