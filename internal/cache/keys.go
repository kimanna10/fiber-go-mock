package cache

import "fmt"

type KeyBuilder struct {
	namespace string
	version   int
}

func NewKeyBuilder(namespace string) *KeyBuilder {
	return &KeyBuilder{
		namespace: namespace,
		version:   1,
	}
}

func (kb *KeyBuilder) BumpVersion() {
	kb.version++
}

// user:{userID}
func (kb *KeyBuilder) User(userID int) string {
	return fmt.Sprintf("%s:v%d:user:%d", kb.namespace, kb.version, userID)
}

// users:page:{pageNumber}:{pageSize}:{userNameFilter}
func (kb *KeyBuilder) Users(page, limit int, name string) string {
	return fmt.Sprintf("%s:v%d:users:page:%d:%d:%s", kb.namespace, kb.version, page, limit, name)
}

// users:*
func (kb *KeyBuilder) UsersPattern() string {
	return fmt.Sprintf("%s:users:*", kb.namespace)
}

// --- НАШИ НОВЫЕ МЕТОДЫ ДЛЯ WEBSOCKET ---

// status:user:{userID} -> например: app:v1:status:user:42
func (kb *KeyBuilder) UserStatus(userID int) string {
	return fmt.Sprintf("%s:v%d:status:user:%d", kb.namespace, kb.version, userID)
}

// status:user:* -> паттерн для поиска всех статусов в Redis
func (kb *KeyBuilder) UserStatusPattern() string {
	return fmt.Sprintf("%s:v%d:status:user:*", kb.namespace, kb.version)
}
