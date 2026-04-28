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
	ErrorDBNotConnect  = errors.New("ОШИБКА: Нет соединения с базой")
	ErrorUserNotFound  = errors.New("ОШИБКА: Пользователь не найден")
	ErrorWrongRows     = errors.New("ОШИБКА: Ошибка при чтении строки")
	ErrorDataUpdate    = errors.New("ОШИБКА: Ошибка при изменении данных")
	ErrorNonexistentID = errors.New("ОШИБКА: Несуществующий ID")
)

type User struct {
	ID    int
	Name  string
	Email string
	Phone string //lox
}

func PrintAllUsersFromDB(db *sql.DB) {
	rows, err := db.Query(`
	SELECT * FROM users
	`)
	if err != nil {
		fmt.Println(ErrorWrongRows)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var ID int
		var name, email, phone string

		rows.Scan(&ID, &name, &email, &phone)
		fmt.Printf("ID: %d, Name: %s, Email: %s, Phone: %s\n", ID, name, email, phone)
	}
}

func DelUserFromDB(db *sql.DB, ID int) {
	res, err := db.Exec(`DELETE FROM users WHERE id = $1`, ID)
	if err != nil {
		fmt.Println(ErrorDBNotConnect)
		return
	}
	rowsAffected, err := res.RowsAffected()
	if rowsAffected == 0 {
		fmt.Println(ErrorUserNotFound)
		return
	}
	fmt.Printf("Пользователь с ID %d удален\n", ID)
}

func UpdateUserFromDB(db *sql.DB, ID int) {
	var name, email, phone string
	fmt.Println("Введите новые данные через пробел (Имя Email Телефон):") 
	fmt.Scan(&name, &email, &phone) //???????????????????????????????????????????????????????????
	res, err := db.Exec(`
        UPDATE users 
        SET name = $1, email = $2, phone = $3 
        WHERE id = $4
    `, name, email, phone, ID)

	if err != nil {
		fmt.Println(ErrorDataUpdate, err)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil{
		fmt.Println(err)
		return
	}
	if rowsAffected > 0 {
		fmt.Printf("Пользователь c ID %d успешно обновлен!\n", ID)
		return
	}
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
		fmt.Println("3. Вывести всех пользователей")
		fmt.Println("4. Изменить пользователя по ID")
		fmt.Println("5. Удалить пользователя по ID")
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

		case "3":
			PrintAllUsersFromDB(db)

		case "4":
			var ID int
			fmt.Printf("Введите ID пользователя, данные которго хотите обновить:")
			fmt.Scan(&ID)
			UpdateUserFromDB(db, ID)

		case "5":
			var ID int
			fmt.Printf("Введите ID пользователя, данные которого хотите удалить:")
			fmt.Scan(&ID)
			DelUserFromDB(db, ID)

		case "0":
			fmt.Println("Завершение работы...")
			return

		default:
			fmt.Println("Неверный ввод, попробуйте еще раз.")
		}
	}
}
