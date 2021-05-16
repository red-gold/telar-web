package models

type CreateSettingGroupModel struct {
	Type string                  `json:"type"`
	List []SettingGroupItemModel `json:"list"`
}

type CreateMultipleSettingsModel struct {
	List []CreateSettingGroupModel `json:"list"`
}
