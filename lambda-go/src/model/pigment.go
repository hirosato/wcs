package model

type Pigment struct {
	Key         int32  `json:"key" db:"key, primarykey"`
	Name        string `json:"name" db:"name"`
}
