package domain

import "github.com/hirosato/wcs/model"

type LocalFileRepository interface {
	Add(userId string, timestamp string, base64image string, imageKind model.ImageKind) (string, error)
	Remove(filename string)
}

type S3Repository interface {
	Add(localFilePath string, userId string, timestamp string, filename model.ImageKind) error
}
