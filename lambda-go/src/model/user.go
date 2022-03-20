package model

type User struct {
	UserId      string `json:"userId" dynamodbav:"UserId"`
	DisplayName string `json:"displayName" dynamodbav:"DisplayName"`
	AvatarURL   string `json:"avatarUrl" dynamodbav:"AvatarURL"`
}
