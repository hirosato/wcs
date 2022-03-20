package model

type ImageKind string

func (imageKind ImageKind) ToPathString() string {
	return string(imageKind)
}

const (
	ImageCover = ImageKind("cover")
	Image1     = ImageKind("1")
	Image2     = ImageKind("2")
	Image3     = ImageKind("3")
	Image4     = ImageKind("4")
)

type PaintingImage struct {
	UserId     string `json:"user_id"`
	Timestamp  string `json:"timestamp"`
	ImageCover string `json:"image_cover"`
	Image1     string `json:"image1"`
	Image2     string `json:"image2"`
	Image3     string `json:"image3"`
	Image4     string `json:"image4"`
}

type Painting struct {
	UserId             string `json:"user_id"`
	Timestamp          string `json:"timestamp"`
	Date               string `json:"date"`
	Title              string `json:"title"`
	Description        string `json:"description"`
	Created            uint64 `json:"created"`
	Updated            uint64 `json:"updated"`
	Likes              uint32 `json:"likes"`
	Favorits           uint32 `json:"favorits"`
	HasImageCover      bool   `json:"has_image_cover"`
	HasImage1          bool   `json:"has_image1"`
	HasImage2          bool   `json:"has_image2"`
	HasImage3          bool   `json:"has_image3"`
	HasImage4          bool   `json:"has_image4"`
}

func (painting *Painting) GetId() string {
	return painting.UserId + "-" + painting.Timestamp
}
