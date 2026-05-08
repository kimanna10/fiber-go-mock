package cache

import "fmt"

type KeyBuilder struct {
	namespace string
}

func NewKeyBuilder(namespace string) *KeyBuilder {
	return &KeyBuilder{namespace: namespace}
}

// user:{userID}
func (kb *KeyBuilder) User(userID int) string {
	return fmt.Sprintf("%s:user:%d", kb.namespace, userID)
}

// users:page:{pageNumber}:{pageSize}:{userNameFilter}
func (kb *KeyBuilder) Users(page, limit int, name string) string {
	return fmt.Sprintf("%s:users:page:%d:%d:%s", kb.namespace, page, limit, name)
}

// users:*
func (kb *KeyBuilder) UsersPattern() string {
	return fmt.Sprintf("%s:users:*", kb.namespace)
}
