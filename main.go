package main

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgtype"
)

func main() {
	db := NewDB(NewDBConfig())

	store := NewImageStore(db)
	f, err := os.Open("./tmp/foo.png")
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	fileStat, err := f.Stat()
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	uploader := NewUploader(NewUploaderConfig())
	res, err := uploader.Upload(ctx, UploadRequest{
		Reader:   f,
		Filename: fileStat.Name(),
	})
	if err != nil {
		log.Fatalln(err)
	}

	var tags pgtype.TextArray
	if err := tags.Set([]string{"hello world", "foo", "bar"}); err != nil {
		log.Fatalln(err)
	}
	img, err := store.Create(ctx, CreateRequest{
		Bucket:  res.Bucket,
		Key:     res.Key,
		Width:   res.Width,
		Height:  res.Height,
		Version: res.VersionID,
		Tags:    tags,
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("%+v", img)

	images, err := store.FindAll(ctx, NewPagination())
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(images)
}
