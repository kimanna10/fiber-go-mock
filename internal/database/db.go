package database

import (
	"database/sql"
	"fiber-go/internal/config" // Импортируем наш конфиг
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// Теперь функция возвращает объект, а не пишет в глобальную переменную
func Connect(cfg config.DBConfig) *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Ошибка открытия базы:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("База недоступна:", err)
	}

	log.Println("Успешное подключение к БД")
	return db
}
