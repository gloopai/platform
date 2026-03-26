package store

import (
	"context"
	"strconv"

	"gorm.io/gorm"
)

const (
	GlobalSettingCountryCode     = "country_code"
	GlobalSettingCurrencyCode    = "currency_code"
	GlobalSettingCurrencySign    = "currency_symbol"
	GlobalSettingFiatToUsdtRate  = "fiat_to_usdt_rate"
	GlobalSettingAdminMfaEnabled = "admin_mfa_enabled"
)

type GlobalDisplaySettings struct {
	CountryCode     string
	CurrencyCode    string
	CurrencySymbol  string
	FiatToUsdtRate  float64
	AdminMfaEnabled int64
}

type GlobalSettingsStore struct {
	db *gorm.DB
}

func NewGlobalSettingsStore(db *gorm.DB) *GlobalSettingsStore {
	return &GlobalSettingsStore{db: db}
}

func (s *GlobalSettingsStore) GetDisplaySettings(ctx context.Context) (*GlobalDisplaySettings, error) {
	type kv struct {
		K string `gorm:"column:setting_key"`
		V string `gorm:"column:setting_value"`
	}
	var rows []kv
	if err := s.db.WithContext(ctx).
		Table("global_settings").
		Select("setting_key, setting_value").
		Where("setting_key IN ?", []string{GlobalSettingCountryCode, GlobalSettingCurrencyCode, GlobalSettingCurrencySign, GlobalSettingFiatToUsdtRate, GlobalSettingAdminMfaEnabled}).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := &GlobalDisplaySettings{
		CountryCode:     "CN",
		CurrencyCode:    "CNY",
		CurrencySymbol:  "¥",
		FiatToUsdtRate:  7.2,
		AdminMfaEnabled: 0,
	}
	for _, r := range rows {
		switch r.K {
		case GlobalSettingCountryCode:
			out.CountryCode = r.V
		case GlobalSettingCurrencyCode:
			out.CurrencyCode = r.V
		case GlobalSettingCurrencySign:
			out.CurrencySymbol = r.V
		case GlobalSettingFiatToUsdtRate:
			if v, err := strconv.ParseFloat(r.V, 64); err == nil && v > 0 {
				out.FiatToUsdtRate = v
			}
		case GlobalSettingAdminMfaEnabled:
			if r.V == "1" {
				out.AdminMfaEnabled = 1
			} else {
				out.AdminMfaEnabled = 0
			}
		}
	}
	return out, nil
}

func (s *GlobalSettingsStore) UpsertDisplaySettings(ctx context.Context, in *GlobalDisplaySettings) error {
	return s.db.WithContext(ctx).Exec(`
INSERT INTO global_settings (setting_key, setting_value) VALUES
  (?, ?),
  (?, ?),
  (?, ?),
  (?, ?),
  (?, ?)
ON DUPLICATE KEY UPDATE setting_value = VALUES(setting_value)
`,
		GlobalSettingCountryCode, in.CountryCode,
		GlobalSettingCurrencyCode, in.CurrencyCode,
		GlobalSettingCurrencySign, in.CurrencySymbol,
		GlobalSettingFiatToUsdtRate, strconv.FormatFloat(in.FiatToUsdtRate, 'f', 6, 64),
		GlobalSettingAdminMfaEnabled, strconv.FormatInt(in.AdminMfaEnabled, 10),
	).Error
}
