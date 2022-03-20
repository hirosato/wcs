package env

import "os"

var IsLocal bool = os.Getenv("IS_LOCAL") == "TRUE"
var bucketName string = os.Getenv("BUCKET_NAME")
var Region = os.Getenv("AWS_REGION")

func init() {
	if Region == "" {
		Region = "ap-northeast-1"
	}
}

func GetFrontUrl() string {
	if IsLocal {
		return "http://localhost:4200"
	} else {
		return "https://watercolor.site"
	}
}
func GetApiUrl() string {
	if IsLocal {
		return "http://localhost:8080"
	} else {
		return "https://api.watercolor.site"
	}
}
func GetEsUrl() string {
	if IsLocal {
		return "https://es.watercolor.site"
		// return "http://localhost:9200"
	} else {
		return "https://es.watercolor.site"
	}
}
func GetBucketName() string {
	return bucketName
}
