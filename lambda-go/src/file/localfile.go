package file

import (
	"errors"
	"os"

	"github.com/hirosato/wcs/domain"
	"github.com/hirosato/wcs/model"
	"github.com/vincent-petithory/dataurl"
)

type localFileRepositoryImpl struct {
	Request *model.Painting
}

func NewLocalFileRepository() domain.LocalFileRepository {
	return &localFileRepositoryImpl{}
}

func (impl *localFileRepositoryImpl) Add(userId string, timestamp string, base64image string, imageKind model.ImageKind) (string, error) {
	dataURL, err := dataurl.DecodeString(base64image)
	if err != nil {
		return "", err
	}
	if dataURL.ContentType() == "image/png" {
		filename := "/tmp/" + userId + "-" + timestamp + "-" + imageKind.ToPathString() + ".png"
		file, err := os.Create(filename)
		if err != nil {
			return filename, err
		}
		defer file.Close()
		file.Write(dataURL.Data)
		return filename, nil
	}
	return "", errors.New("something went wrong")
}

//ignore remove failure since its on lambda anyway.
func (impl *localFileRepositoryImpl) Remove(filename string) {
	if fileExists(filename) {
		os.Remove(filename)
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
