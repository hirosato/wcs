package file

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hirosato/wcs/domain"
	"github.com/hirosato/wcs/env"
	"github.com/hirosato/wcs/model"
)

type s3RepositoryImpl struct{}

func NewS3RepositoryImpl() domain.S3Repository {
	return s3RepositoryImpl{}
}

func (impl s3RepositoryImpl) Add(localFilePath string, userId string, timestamp string, filename model.ImageKind) error {
	file, openErr := os.Open(localFilePath)
	if openErr != nil {
		return openErr
	}
	defer file.Close()

	if uploadErr := impl.upload(file, userId, timestamp, filename); uploadErr != nil {
		return uploadErr
	}

	return nil
}

func (impl s3RepositoryImpl) upload(file *os.File, userId string, timestamp string, filename model.ImageKind) error {
	uploader := s3manager.NewUploader(impl.newSession())

	if _, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(env.GetBucketName()),
		Key:    aws.String("/wcs/" + userId + "/" + timestamp + "/" + filename.ToPathString() + ".png"),
		Body:   file,
	}); err != nil {
		return err
	}

	return nil
}

func (impl s3RepositoryImpl) newSession() *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-1"),
	}))
}
