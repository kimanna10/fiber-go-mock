package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"fiber-go/internal/cache"
	"fiber-go/internal/errs"
	"fiber-go/internal/models"
	"fiber-go/internal/patterns"
	"fiber-go/internal/repository"

	"github.com/hibiken/asynq"
	"golang.org/x/sync/singleflight"
)

// 1. Описываем интерфейс со всеми методами
type UserService interface {
	GetUsers(ctx context.Context, limit, page int, name string) ([]models.UserResponse, error)
	GetUserById(ctx context.Context, id int) (models.UserResponse, error)
	Register(ctx context.Context, req models.UserRegisterRequest) (models.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	// Update с map
	// UpdateUser(ctx context.Context, id int, data map[string]interface{}) (models.UserResponse, error)

	// Update c patch структурой
	UpdateUser(ctx context.Context, id int, req models.UserUpdateRequest) (models.UserResponse, error)
	DeleteUser(ctx context.Context, id int) error
}

// Делаем структуру приватной (с маленькой буквы)
type userService struct {
	repo repository.UserRepository // Используем интерфейс репозитория
	// cache map[int]models.User
	// mu    sync.RWMutex
	cache cache.Cache
	keys  *cache.KeyBuilder

	asynqClient *asynq.Client
	group       singleflight.Group
}

// NewUserService внедряет зависимость репозитория в сервис
func NewUserService(repo repository.UserRepository, cache cache.Cache, keys *cache.KeyBuilder, asynqClient *asynq.Client) UserService {
	return &userService{
		repo: repo,
		// cache: make(map[int]models.User),
		cache:       cache,
		keys:        keys,
		asynqClient: asynqClient,
	}
}

// GetUsers возвращает список UserResponse (DTO)
func (s *userService) GetUsers(ctx context.Context, limit, page int, name string) ([]models.UserResponse, error) {

	// Формируем ключ для кэша
	key := s.keys.Users(page, limit, name)

	// Пытаемся достать из кэша
	cached, err := s.cache.Get(ctx, key)
	if err == nil {
		var resp []models.UserResponse
		if json.Unmarshal(cached, &resp) == nil {
			return resp, nil
		}
	}

	// Реализуем пагинацию
	offset := (page - 1) * limit
	users, err := s.repo.GetUsers(ctx, limit, offset, name)
	if err != nil {
		return nil, err
	}

	// Конвертируем []models.User -> []models.UserResponse
	resp := make([]models.UserResponse, 0, len(users))
	for _, u := range users {
		resp = append(resp, s.mapToResponse(u))
	}

	// Сохраняем в кэш
	data, _ := json.Marshal(resp)
	_ = s.cache.Set(ctx, key, data, 5*time.Minute)

	return resp, nil
}

// GetUserById использует кэш и возвращает UserResponse
func (s *userService) GetUserById(ctx context.Context, id int) (models.UserResponse, error) {
	// s.mu.RLock()
	// cached, ok := s.cache[id]
	// s.mu.RUnlock()
	// if ok {
	// 	return s.mapToResponse(cached), nil
	// }

	// Формируем ключ для кэша
	key := s.keys.User(id) // namespace:user:1

	// Пытаемся достать из кэша
	cached, err := s.cache.Get(ctx, key)
	if err == nil {
		var user models.User
		if json.Unmarshal(cached, &user) == nil {
			return s.mapToResponse(user), nil
		}
	}

	v, err, _ := s.group.Do(key, func() (interface{}, error) {
		// Если в кэше нет, то достаем из БД
		user, err := s.repo.GetUserById(ctx, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errs.ErrUserNotFound
			}
			return nil, err
		}
		data, _ := json.Marshal(user)
		_ = s.cache.Set(ctx, key, data, 5*time.Minute)

		return user, nil
	})

	if err != nil {
		return models.UserResponse{}, err
	}

	// s.mu.Lock()
	// s.cache[id] = user
	// s.mu.Unlock()

	return s.mapToResponse(v.(models.User)), nil
}

// Register (Create) принимает UserRegisterRequest и возвращает UserResponse
// func (s *userService) Register(ctx context.Context, req models.UserRegisterRequest) (models.UserResponse, error) {
// 	if req.Email == "" || req.Password == "" {
// 		return models.UserResponse{}, errs.ErrBadRequest
// 	}

// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		return models.UserResponse{}, err
// 	}

// 	// Собираем модель для БД
// 	user := models.User{
// 		Name:     req.Name,
// 		Age:      req.Age,
// 		Email:    req.Email,
// 		Password: string(hashedPassword),
// 		Role:     "user", // Дефолтная роль
// 	}

// 	if err := s.repo.Create(ctx, &user); err != nil {
// 		return models.UserResponse{}, err
// 	}

// 	return s.mapToResponse(user), nil
// }

func (s *userService) Register(ctx context.Context, req models.UserRegisterRequest) (models.UserResponse, error) {
	// Собираем модель для БД
	user, err := patterns.NewUserBuilder().
		SetName(req.Name).
		SetAge(req.Age).
		SetEmail(req.Email).
		SetPassword(req.Password).
		DefaultRole().
		Build()

	if err != nil {
		return models.UserResponse{}, err
	}
	if err := s.repo.Create(ctx, &user); err != nil {
		return models.UserResponse{}, err
	}

	payload, _ := json.Marshal(map[string]string{
		"Email": user.Email,
		"Name":  user.Name,
	})
	_, _ = s.asynqClient.EnqueueContext(ctx, asynq.NewTask("email:welcome", payload))

	// Удаление из кэша, так как нового пользователя нет в кэше
	// pattern := s.keys.UsersPattern()
	// _ = s.cache.DeleteByPattern(ctx, pattern)

	s.keys.BumpVersion()

	return s.mapToResponse(user), nil
}

// GetUserByEmail — обычно используется для логина,
func (s *userService) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errs.ErrUserNotFound
		}
		return models.User{}, err
	}
	return user, nil
}

// UpdateUser использует мапу для частичного обновления (Patch)
// func (s *userService) UpdateUser(ctx context.Context, id int, data map[string]interface{}) (models.UserResponse, error) {
// 	// Сначала получаем текущего юзера из базы
// 	user, err := s.repo.GetUserById(ctx, id)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			return models.UserResponse{}, errs.ErrUserNotFound
// 		}
// 		return models.UserResponse{}, err
// 	}

// 	// Частично обновляем поля
// 	if name, ok := data["name"].(string); ok {
// 		user.Name = name
// 	}
// 	if age, ok := data["age"].(float64); ok {
// 		user.Age = int(age)
// 	}
// 	if email, ok := data["email"].(string); ok {
// 		user.Email = email
// 	}

// 	if err := s.repo.Update(ctx, &user); err != nil {
// 		return models.UserResponse{}, err
// 	}

// 	// Сбрасываем кэш, так как данные изменились
// 	s.mu.Lock()
// 	delete(s.cache, id)
// 	s.mu.Unlock()

// 	return s.mapToResponse(user), nil
// }

// UpdateUser использует патч структуру для частичного обновления (Patch)
func (s *userService) UpdateUser(ctx context.Context, id int, req models.UserUpdateRequest) (models.UserResponse, error) {
	// Сначала получаем текущего юзера из базы
	user, err := s.repo.GetUserById(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.UserResponse{}, errs.ErrUserNotFound
		}
		return models.UserResponse{}, err
	}

	// Частично обновляем поля
	builder := patterns.NewUserUpdateBuilder(user)
	builder, err = builder.ApplyPatch(req)
	if err != nil {
		return models.UserResponse{}, err
	}
	updatedUser, err := builder.Build()
	if err != nil {
		return models.UserResponse{}, err
	}

	// Сохраняем
	if err := s.repo.Update(ctx, &updatedUser); err != nil {
		return models.UserResponse{}, err
	}

	// Удаляем из кэша, так как данные изменились
	key := s.keys.User(id)
	_ = s.cache.Delete(ctx, key)

	// Удаление из кэша, где мог был присутствовать данный пользователь
	// pattern := s.keys.UsersPattern()
	// _ = s.cache.DeleteByPattern(ctx, pattern)
	s.keys.BumpVersion()

	// Сбрасываем кэш, так как данные изменились
	// s.mu.Lock()
	// delete(s.cache, id)
	// s.mu.Unlock()

	return s.mapToResponse(updatedUser), nil
}

func (s *userService) DeleteUser(ctx context.Context, id int) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errs.ErrUserNotFound
		}
		return err
	}

	// s.mu.Lock()
	// delete(s.cache, id)
	// s.mu.Unlock()

	// Удаляем из кэша, так как данные изменились
	key := s.keys.User(id)
	_ = s.cache.Delete(ctx, key)

	// Удаление из кэша, где мог был присутствовать данный пользователь
	// pattern := s.keys.UsersPattern()
	// _ = s.cache.DeleteByPattern(ctx, pattern)
	s.keys.BumpVersion()

	return nil
}

// Вспомогательный метод для маппинга (чтобы не дублировать код)
func (s *userService) mapToResponse(u models.User) models.UserResponse {
	return models.UserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Age:   u.Age,
		Email: u.Email,
		Role:  u.Role,
	}
}
