package models

type Item struct {
	ID          int32  `gorm:"primaryKey"`
	Name        string `gorm:"type:varchar(100)"`
	Price       int32  `gorm:"type:int"`
	Description string `gorm:"type:varchar(255)"`
	Quantity    int32  `gorm:"type:int"`
}
