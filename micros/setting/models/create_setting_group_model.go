package models

type CreateSettingGroupModel struct {
	Type string                  `json:"type"`
	List []SettingGroupItemModel `json:"list"`
}
