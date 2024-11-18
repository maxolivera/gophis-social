package models

type RoleType string

const (
	RoleAdmin     RoleType = RoleType("admin")
	RoleModerator RoleType = RoleType("moderator")
	RoleUser      RoleType = RoleType("user")
)

// Special Type which do not contains Description nor ID
type ReducedRole struct {
	Name  RoleType `json:"name"`
	Level int      `json:"level"`
}
