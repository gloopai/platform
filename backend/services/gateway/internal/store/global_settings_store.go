package store

import (
	"context"
	"database/sql"
)

const (
	GlobalSettingCountryCode  = "country_code"
	GlobalSettingCurrencyCode = "currency_code"
	GlobalSettingCurrencySign = "currency_symbol"
)

type GlobalDisplaySettings struct {
	CountryCode    string
	CurrencyCode   string
	CurrencySymbol string
}

type GlobalSettingsStore struct {
	db *sql.DB
}

func NewGlobalSettingsStore(db *sql.DB) *GlobalSettingsStore {
	return &GlobalSettingsStore{db: db}
}

func (s *GlobalSettingsStore) GetDisplaySettings(ctx context.Context) (*GlobalDisplaySettings, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT setting_key, setting_value
FROM global_settings
WHERE setting_key IN (?, ?, ?)
`, GlobalSettingCountryCode, GlobalSettingCurrencyCode, GlobalSettingCurrencySign)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := &GlobalDisplaySettings{
		CountryCode:    "CN",
		CurrencyCode:   "CNY",
		CurrencySymbol: "¥",
	}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		switch k {
		case GlobalSettingCountryCode:
			out.CountryCode = v
		case GlobalSettingCurrencyCode:
			out.CurrencyCode = v
		case GlobalSettingCurrencySign:
			out.CurrencySymbol = v
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *GlobalSettingsStore) UpsertDisplaySettings(ctx context.Context, in *GlobalDisplaySettings) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO global_settings (setting_key, setting_value) VALUES
  (?, ?),
  (?, ?),
  (?, ?)
ON DUPLICATE KEY UPDATE setting_value = VALUES(setting_value)
`,
		GlobalSettingCountryCode, in.CountryCode,
		GlobalSettingCurrencyCode, in.CurrencyCode,
		GlobalSettingCurrencySign, in.CurrencySymbol,
	)
	return err
}
