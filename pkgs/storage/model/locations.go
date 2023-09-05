package model

type Location struct {
	*Model
	IP      string `json:"ip"`
	Country string `json:"country"`
}

func (that Location) TableName() string {
	return "locations"
}
