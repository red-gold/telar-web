package constants

import "strings"

type ActionStatusConst int

const (
	ActionStatusConstConnect ActionStatusConst = iota
	ActionStatusConstIdle
	ActionStatusConstDisconnect
)

func (a *ActionStatusConst) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)

	switch {
	case str == "Connect":
		*a = ActionStatusConstConnect
	case str == "Idle":
		*a = ActionStatusConstIdle
	case str == "Disconnect":
		*a = ActionStatusConstDisconnect

	default:
		*a = ActionStatusConstDisconnect
	}

	return nil
}
