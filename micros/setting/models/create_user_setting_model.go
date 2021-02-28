package models

type CreateUserSettingModel struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type"`
}
