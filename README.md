# MinIO


```bash
$ brew install minio/stable/mc

# Set local environment to point to docker.
$ mc alias set local http://localhost:900 minio minio123

# Create a bucket called `test` using the `local` environment.
$ mc mb local/test

# List all buckets.
$ mc ls local
```

## Creating user and group


```bash
# Create a new user `newuser` with secret key `newuser123`. The secret key must be between 8 and 40 characters.
$ mc admin user add local newuser newuser123

# Attach user to the existing `readwrite` policy.
$ mc admin policy set local readwrite user=newuser

# Create a new group.
$ mc admin group add local newgroup newuser

# Apply `readwrite` policy to the group.
$ mc admin policy set local readwrite group=newgroup
````

References:
- https://docs.min.io/docs/minio-multi-user-quickstart-guide.html

## Allowing only certain image type.

Create the following policy

```json
{
  "Version": "2012-10-17",
  "Id": "Policy1464968545158",
  "Statement": [
    {
      "Sid": "Stmt1464968483619",
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::111111111111:user/exampleuser"
      },
      "Action": "s3:PutObject",
      "Resource": [
        "arn:aws:s3:::DOC-EXAMPLE-BUCKET/*.jpg",
        "arn:aws:s3:::DOC-EXAMPLE-BUCKET/*.png",
        "arn:aws:s3:::DOC-EXAMPLE-BUCKET/*.gif"
      ]
    },
    {
      "Sid": "Stmt1464968483619",
      "Effect": "Deny",
      "Principal": "*",
      "Action": "s3:PutObject",
      "NotResource": [
        "arn:aws:s3:::DOC-EXAMPLE-BUCKET/*.jpg",
        "arn:aws:s3:::DOC-EXAMPLE-BUCKET/*.png",
        "arn:aws:s3:::DOC-EXAMPLE-BUCKET/*.gif"
      ]
    }
  ]
}
```
- https://aws.amazon.com/premiumsupport/knowledge-center/s3-allow-certain-file-types/


## Database schema

```sql
-- TODO: Split into image directory, and images. We need a pointer to the image
-- directory that returns multiple image resolution.
-- For simplicity, just use the id as the key, id.png. To group a bunch of images, just use another table to store the path mapping to virtual directory.
CREATE TABLE IF NOT EXISTS images (
	id uuid DEFAULT gen_random_uuid(),
	bucket text NOT NULL,
	key text NOT NULL,
	width int NOT NULL,
	height int NOT NULL,
	version text NOT NULL,
	extension text NOT NULL,
	meta jsonb NOT NULL DEFAULT '{}',
	tags text[] NOT NULL DEFAULT '{}',
	created_at timestamptz NOT NULL DEFAULT current_timestamp,
	updated_at timestamptz NOT NULL DEFAULT current_timestamp,
	PRIMARY KEY (id),
	UNIQUE (bucket, key, width, extension)
);

CREATE EXTENSION moddatetime;
CREATE TRIGGER mdt_images
	BEFORE UPDATE ON images
	FOR EACH ROW
	EXECUTE PROCEDURE moddatetime(updated_at);

CREATE INDEX idx_tags ON images USING GIN(tags); -- GIN Index (array)

INSERT INTO images(bucket, key, width, height, extension, version, tags)
VALUES ('mybucket', 'path/to/file', 320, 480, '.png', 'xytz', '{hello}')
ON CONFLICT (bucket, key, width, extension) DO UPDATE SET version = EXCLUDED.version;
```


## Manage images

- how to ensure that images returned are responsive? Most application only assumes one image will be returned. When viewing on mobile and web, we might want to consider returning images with different sizes.
- instead of a single image src, e.g. `foo.png`, we create a collection called `foo/` and place all the images with different resolutions inside, e.g. `foo/320w.png, foo/420w.jpg`. The explanation for `w` descriptor can be found [here](https://stackoverflow.com/questions/40890825/explain-w-in-srcset-of-image)
- image resolution and art direction (not implemented) requires different way of handling on the Frontend
- resizing and compressing. It is important to have standardize sizes, so 320px instead of 277px etc. If I upload file with size 499px, I should be provided an option to resize them to 480px, 360px, 320px etc. Image manipulation is not so important here. When resizing, the aspect ratio should be kept the same.
- how to deal uploading images with different extension? Format them all into a standardize format.
- viewing images stored in S3 and sync to database. However, the syncing will be out of tune once some operation is done in S3/Postgres directly without the other knowing about it. Using S3 Event notification will work, but requires more cost/effort to implement.
- uploading different image sizes for a particular asset. Say if I want to upload an image named `foo`, I should be able to upload all the different sizes and have them previewed/returned as img `srcset`.
- images should be load lazily with the `loading=lazy` HTML attribute.
- bulk tagging and adding metadata for images
- bulk upload images with different resolutions
- how to reference the images stored here from other services? Store the image id (?) can be tricky, since there will be more than one collection of image with different image resolutions.
