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
CREATE TABLE IF NOT EXISTS images (
	id uuid DEFAULT gen_random_uuid(),
	bucket text NOT NULL,
	prefix text NOT NULL DEFAULT '',
	extension text NOT NULL,
	uploaded boolean NOT NULL DEFAULT false,
	created_at timestamptz NOT NULL DEFAULT current_timestamp,
	updated_at timestamptz NOT NULL DEFAULT current_timestamp
);

INSERT INTO images (bucket, prefix, extension) VALUES
('mybucket', 'assets', '.png');
table iamges;
drop table images;

CREATE EXTENSION moddatetime;
CREATE TRIGGER mdt_images
	BEFORE UPDATE ON images
	FOR EACH ROW
	EXECUTE PROCEDURE moddatetime(updated_at);
```
