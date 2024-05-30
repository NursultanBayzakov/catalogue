package models

type Item struct {
	ID          int32  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Price       int32  `json:"price,omitempty"`
	Description string `json:"description,omitempty"`
	Quantity    int32  `json:"quantity,omitempty"`
	ImageURL    string `json:"image_url"`
}
