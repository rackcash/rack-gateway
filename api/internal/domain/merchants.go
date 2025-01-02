package domain

type Merchants struct {
	Model
	ID           uint   `gorm:"primaryKey"`
	MerchantName string `gorm:"unique:size:32;not null"`
	MerchantID   string `gorm:"unique:size:36;not null"`
	ApiKey       string `gorm:"size:64;not null"`
}
