package model

// PayinProductAdmin is one row from payin_products (admin CRUD).
type PayinProductAdmin struct {
	ID            int64  `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Code          string `json:"code,omitempty" gorm:"column:code"`
	Name          string `json:"name,omitempty" gorm:"column:name"`
	SortOrder     int64  `json:"sort_order,omitempty" gorm:"column:sort_order"`
	Enabled       bool   `json:"enabled,omitempty" gorm:"column:enabled"`
	ProductConfig string `json:"product_config,omitempty" gorm:"column:product_config"`
}

// PayinProductBindingAdmin is payin_product_channels plus joined channels.name.
type PayinProductBindingAdmin struct {
	ID             int64  `json:"id,omitempty" gorm:"column:id;primaryKey"`
	PayinProductID int64  `json:"payin_product_id,omitempty" gorm:"column:payin_product_id"`
	ChannelID      int64  `json:"channel_id,omitempty" gorm:"column:channel_id"`
	ChannelName    string `json:"channel_name,omitempty" gorm:"column:channel_name"`
	Weight         int64  `json:"weight,omitempty" gorm:"column:weight"`
	Enabled        bool   `json:"enabled,omitempty" gorm:"column:enabled"`
}

// PayoutProductAdmin is one row from payout_products (admin CRUD).
type PayoutProductAdmin struct {
	ID            int64  `json:"id,omitempty" gorm:"column:id;primaryKey"`
	Code          string `json:"code,omitempty" gorm:"column:code"`
	Name          string `json:"name,omitempty" gorm:"column:name"`
	SortOrder     int64  `json:"sort_order,omitempty" gorm:"column:sort_order"`
	Enabled       bool   `json:"enabled,omitempty" gorm:"column:enabled"`
	ProductConfig string `json:"product_config,omitempty" gorm:"column:product_config"`
}

// PayoutProductBindingAdmin is payout_product_channels plus joined channels.name.
type PayoutProductBindingAdmin struct {
	ID              int64  `json:"id,omitempty" gorm:"column:id;primaryKey"`
	PayoutProductID int64  `json:"payout_product_id,omitempty" gorm:"column:payout_product_id"`
	ChannelID       int64  `json:"channel_id,omitempty" gorm:"column:channel_id"`
	ChannelName     string `json:"channel_name,omitempty" gorm:"column:channel_name"`
	Weight          int64  `json:"weight,omitempty" gorm:"column:weight"`
	Enabled         bool   `json:"enabled,omitempty" gorm:"column:enabled"`
}

// PayinProductOption is a minimal code/name pair for checkout / terminal listing.
type PayinProductOption struct {
	Code string `json:"code,omitempty"`
	Name string `json:"name,omitempty"`
}
