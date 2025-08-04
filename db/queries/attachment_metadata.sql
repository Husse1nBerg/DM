-- name: CreateAttachmentMetadata :one
INSERT INTO dme_attachment_metadata (
    s3_path,
    public
) VALUES (
    $1, $2
) RETURNING id, s3_path, public;

-- name: GetAttachmentMetadataByS3Path :one
SELECT id, s3_path, public
FROM dme_attachment_metadata
WHERE s3_path = $1;

-- name: UpdateAttachmentMetadataPublic :one
UPDATE dme_attachment_metadata
SET public = $2
WHERE s3_path = $1
RETURNING id, s3_path, public;

-- name: ListAttachmentMetadata :many
SELECT id, s3_path, public
FROM dme_attachment_metadata;
