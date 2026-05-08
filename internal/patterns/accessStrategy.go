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

// OwnerOrAdminStrategy позволяет владельцу данных или администратору получить доступ
type OwnerOrAdminStrategy struct{}

func OwnerOrAdmin() AccessStrategy {
	return &OwnerOrAdminStrategy{}
}

func (s *OwnerOrAdminStrategy) CanAccess(ctx AccessContext) bool {
	return ctx.UserId == ctx.TargetId || ctx.Role == RoleAdmin
}

// OnlyOwnerStrategy позволяет получить доступ только владельцу данных
type OnlyOwnerStrategy struct{}

func OnlyOwner() AccessStrategy {
	return &OnlyOwnerStrategy{}
}

func (s *OnlyOwnerStrategy) CanAccess(ctx AccessContext) bool {
	return ctx.UserId == ctx.TargetId
}

// OnlyAdminStrategy позволяет получить доступ только администратору
type OnlyAdminStrategy struct{}

func OnlyAdmin() AccessStrategy {
	return &OnlyAdminStrategy{}
}

func (s *OnlyAdminStrategy) CanAccess(ctx AccessContext) bool {
	return ctx.Role == RoleAdmin
}
