package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq" // Драйвер PostgreSQL
)

var (
	ErrorDBNotConnect = errors.New("Нет соединения с базой")
	ErrorUserNotFound = errors.New("Пользователь не найден")
	ErrorWrongRows    = errors.New("Ошибка при чтении строки")
)

type User struct {
	ID    int
	Name  string
	Email string
	Phone string
}

func main() {
	// 1. Получаем настройки из переменных окружения (Docker Compose)
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME") // Исправлено: было DBNAME

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// 2. Подключение с небольшим ожиданием (Retry logic)
	var db *sql.DB
	var err error

	fmt.Println("⏳ Ожидание запуска базы данных...")
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		fmt.Printf("Попытка %d: База еще не готова, ждем...\n", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		fmt.Printf("❌ Не удалось подключиться к базе после 5 попыток: %v\n", err)
		return
	}
	defer db.Close()

	fmt.Println("✅ Успешное подключение к PostgreSQL!")

	// 3. Инициализация таблицы
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  name TEXT,
  email TEXT,
  phone TEXT
 )`)
	if err != nil {
		fmt.Printf("Ошибка при создании таблицы: %v\n", err)
	}

	// 4. Основной цикл программы
	for {
		fmt.Println("\n=== МЕНЮ ===")
		fmt.Println("1. Зарегистрироваться")
		fmt.Println("2. Найти пользователя по ID")
		fmt.Println("0. Выйти")
		fmt.Print("Выберите действие: ")

		var choice string
		fmt.Scan(&choice)

		switch choice {
		case "1":
			var u User
			fmt.Println("Введите данные через пробел (Имя Email Телефон):")
			fmt.Scan(&u.Name, &u.Email, &u.Phone)

			err := db.QueryRow(
				"INSERT INTO users (name, email, phone) VALUES($1, $2, $3) RETURNING id",
				u.Name, u.Email, u.Phone,
			).Scan(&u.ID)

			if err != nil {
				fmt.Println("Ошибка при добавлении:", err)
			} else {
				fmt.Printf("🎉 Пользователь добавлен! Присвоен ID: %d\n", u.ID)
			}

		case "2":
			var id int
			fmt.Print("Введите ID: ")
			fmt.Scan(&id)

			var u User
			err := db.QueryRow("SELECT id, name, email, phone FROM users WHERE id = $1", id).
				Scan(&u.ID, &u.Name, &u.Email, &u.Phone)

			if err == sql.ErrNoRows {
				fmt.Println("❌ Пользователь не найден")
			} else if err != nil {
				fmt.Println("Ошибка запроса:", err)
			} else {
				fmt.Printf("👤 Найден: ID: %d | Имя: %s | Email: %s | Тел: %s\n", u.ID, u.Name, u.Email, u.Phone)
			}

		case "0":
			fmt.Println("Завершение работы...")
			return

		default:
			fmt.Println("Неверный ввод, попробуйте еще раз.")
		}
	}
}
