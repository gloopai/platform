package model

// Merchant is a row from table merchants.
type Merchant struct {
	ID               int64  `json:"id,omitempty" gorm:"column:id;primaryKey"`
	MerchantId       string `json:"merchant_id,omitempty" gorm:"column:merchant_id"`
	AppId            string `json:"app_id,omitempty" gorm:"column:app_id"`
	Email            string `json:"email,omitempty" gorm:"column:email"`
	AppSecret        string `json:"app_secret,omitempty" gorm:"column:app_secret"`
	PasswordHash     string `json:"-" gorm:"column:password_hash"`
	Status           int64  `json:"status,omitempty" gorm:"column:status"`
	IpWhitelist      string `json:"ip_whitelist,omitempty" gorm:"column:ip_whitelist"`
	PayinBalance     int64  `json:"payin_balance,omitempty" gorm:"column:payin_balance"`
	AvailableBalance int64  `json:"available_balance,omitempty" gorm:"column:available_balance"`
	FrozenBalance    int64  `json:"frozen_balance,omitempty" gorm:"column:frozen_balance"`
	WithdrawnAmount  int64  `json:"withdrawn_amount,omitempty" gorm:"column:withdrawn_amount"`
	NotifyUrl        string `json:"notify_url,omitempty" gorm:"column:notify_url"`
	ReturnUrl        string `json:"return_url,omitempty" gorm:"column:return_url"`
	MerchantConfig   string `json:"merchant_config,omitempty" gorm:"column:merchant_config"`
}
