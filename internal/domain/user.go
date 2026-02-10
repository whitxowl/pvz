package domain

type Role string

const (
	RoleModerator Role = "moderator"
	RoleEmployee  Role = "employee"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleModerator, RoleEmployee:
		return true
	default:
		return false
	}
}

type User struct {
	ID       string
	Email    string
	PassHash string
	Role     Role
}

type TokenClaims struct {
	Role Role
}
