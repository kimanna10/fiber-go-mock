package patterns

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type AccessContext struct {
	UserId   int
	TargetId int
	Role     Role
}

type AccessStrategy interface {
	CanAccess(ctx AccessContext) bool
}

type OwnerOrAdminStrategy struct {
}

func OwnerOrAdmin() AccessStrategy {
	return &OwnerOrAdminStrategy{}
}

func (s *OwnerOrAdminStrategy) CanAccess(ctx AccessContext) bool {
	return ctx.UserId == ctx.TargetId || ctx.Role == RoleAdmin
}
