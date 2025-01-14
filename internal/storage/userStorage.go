package storage

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

// UserStorageInterface определяет методы для работы с хранилищем пользователей.
type UserStorageInterface interface {
	// IsCookieExist проверяет, существует ли куки пользователя
	// Возвращает true, если куки существует, и false в противном случае.
	IsCookieExist(uniqueID string) bool

	// GetUserID проверяет, существует ли пользователь по указанному уникальному идентификатору.
	// Возвращает true, если пользователь существует, и false в противном случае.
	GetUserID(login string) (int, error)

	// SaveUserCookie сохраняет куки пользователя
	// Возвращает ошибку, если сохранение не удалось.
	SaveUserCookie(cookieValue string, userID int) error

	// UpdateUserCookie обновляет куки пользователя
	// Возвращает ошибку, если сохранение не удалось.
	UpdateUserCookie(cookieValue string, userID int) error

	// SaveUser сохраняет пользователя
	// Возвращает ошибку, если сохранение не удалось.
	SaveUser(login string, password string) (int, error)

	// SaveUserToken сохраняет токен пользователя
	// Возвращает ошибку, если сохранение не удалось.
	SaveUserToken(login string, password string) (int, error)

	// CheckUserCredentials проверяет валидность данных аутентификации
	CheckUserCredentials(login string, password string) (int, error)

	// Init инициализирует хранилище пользователей с помощью строки соединения.
	// Возвращает ошибку, если произошла ошибка инициализации.
	Init(connectionString string) error

	// Close закрывает соединение с хранилищем данных.
	Close()

	// GetUsersCount получение количества пользователей
	GetUsersCount() (int, error)

	// Ping проверяет состояние соединения с базой данных.
	Ping() bool

	// BeginTransaction получение объекта транзакции
	BeginTransaction() error

	// Rollback откатывает текущую транзакцию.
	// Возвращает ошибку, если произошла ошибка при откате.
	Rollback() error

	// Commit фиксирует текущую транзакцию.
	// Возвращает ошибку, если произошла ошибка при фиксации.
	Commit() error

	// AddData запись данных в хранилище
	AddData(userID int, body []byte, dataType string) error
}

// UsersStorage предоставляет реализацию для работы с хранилищем пользователей
// и взаимодействия с базой данных через пул соединений pgx.
type UsersStorage struct {
	// Conn представляет пул соединений с базой данных, позволяющий выполнять SQL-команды и запросы.
	Conn *pgxpool.Pool

	// Ctx представляет контекст, используемый для управления временем жизни запросов и операций.
	Ctx         context.Context
	Transaction pgx.Tx
}

// IsCookieExist проверяет, существует ли куки пользователя
// Возвращает true, если куки существует, и false в противном случае.
func (us *UsersStorage) IsCookieExist(uniqueID string) bool {
	query := "SELECT id FROM users_cookie WHERE user_id = $1"
	rows, err := us.Conn.Query(us.Ctx, query, uniqueID)

	if err != nil {
		return false
	}

	var id int
	var rowsCount int

	for rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return false
		}

		rowsCount++
	}

	return rowsCount > 0
}

// GetUserID проверяет, существует ли пользователь по его уникальному идентификатору.
// Возвращает true, если пользователь существует, и false в противном случае.
func (us *UsersStorage) GetUserID(login string) (int, error) {
	query := "SELECT id FROM users WHERE login = $1"
	rows, err := us.Conn.Query(us.Ctx, query, login)

	if err != nil {
		return 0, err
	}

	var id int
	var rowsCount int

	for rows.Next() {
		if err = rows.Scan(&id); err != nil {
			return 0, err
		}

		rowsCount++
	}

	if rowsCount == 0 {
		return 0, errors.New("пользователь не найден")
	}

	return id, nil
}

// SaveUserCookie сохраняет куки пользователя с указанным уникальным идентификатором.
// Возвращает ошибку, если сохранение не удалось.
func (us *UsersStorage) SaveUserCookie(cookieValue string, userID int) error {
	query := "INSERT INTO user_cookies (user_id, cookie_value) VALUES ($1, $2)"
	_, err := us.Conn.Exec(us.Ctx, query, userID, cookieValue)
	return err
}

// UpdateUserCookie обновляет куки пользователя с указанным уникальным идентификатором.
// Возвращает ошибку, если сохранение не удалось.
func (us *UsersStorage) UpdateUserCookie(cookieValue string, userID int) error {
	query := "UPDATE user_cookies SET cookie_value = $1 where user_id = $2"
	_, err := us.Conn.Exec(us.Ctx, query, cookieValue, userID)
	return err
}

// SaveUser сохраняет нового пользователя с указанным уникальным идентификатором.
// Возвращает ошибку, если сохранение не удалось.
func (us *UsersStorage) SaveUser(login string, password string) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return 0, err // Возвращаем ошибку, если не удалось хешировать
	}

	var userID int
	query := "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id"

	err = us.Transaction.QueryRow(us.Ctx, query, login, hashedPassword).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// SaveUserToken сохраняет токен пользователя
// Возвращает ошибку, если сохранение не удалось.
func (us *UsersStorage) SaveUserToken(login string, password string) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return 0, err // Возвращаем ошибку, если не удалось хешировать
	}

	var userID int
	query := "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id"

	err = us.Transaction.QueryRow(us.Ctx, query, login, hashedPassword).Scan(&userID)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// CheckUserCredentials проверяет валидность данных аутентификации
func (us *UsersStorage) CheckUserCredentials(login string, password string) (int, error) {
	query := "SELECT id, password FROM users where login = $1"

	rows, err := us.Conn.Query(us.Ctx, query, login)

	if err != nil {
		return 0, err
	}

	var userID int
	var hashedPassword string
	var rowsCount int

	for rows.Next() {
		if err = rows.Scan(&userID, &hashedPassword); err != nil {
			return 0, err
		}

		rowsCount++
	}

	if rowsCount == 0 {
		return 0, errors.New("пользователь не найден")
	}

	if bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) != nil {
		return 0, errors.New("ошибка аутентификации")
	}

	return userID, nil
}

// Init инициализирует соединение с базой данных по заданной строке подключения.
// Параметры:
//   - connectionString: строка подключения к базе данных.
//
// Возвращает ошибку, если инициализация соединения не удалась.
func (us *UsersStorage) Init(connectionString string) error {
	us.Ctx = context.Background()
	var err error
	us.Conn, err = pgxpool.Connect(us.Ctx, connectionString)

	if err != nil {
		log.Fatalf("Error while initializing db connection: %v", err)
	}

	return nil
}

// Close закрывает соединение с базой данных.
// Этот метод должен вызываться для освобождения всех ресурсов, занимаемых соединением.
func (us *UsersStorage) Close() {
	us.Conn.Close()
}

// GetUsersCount получение количества пользователей
func (us *UsersStorage) GetUsersCount() (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM users"

	row := us.Conn.QueryRow(us.Ctx, query)

	if err := row.Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// Ping проверяет состояние соединения с базой данных.
// Возвращает true, если соединение успешно, и false, если возникает ошибка.
func (us *UsersStorage) Ping() bool {
	if err := us.Conn.Ping(us.Ctx); err != nil {
		return false
	}

	return true
}

// BeginTransaction получение объекта транзакции
func (us *UsersStorage) BeginTransaction() error {
	var err error
	us.Transaction, err = us.Conn.BeginTx(
		us.Ctx, pgx.TxOptions{})

	return err
}

// Rollback откатывает текущую транзакцию.
// Возвращает ошибку, если произошла ошибка при откате.
func (us *UsersStorage) Rollback() error {
	if us.Transaction == nil {
		return errors.New("транзакция не инициализирована")
	}

	return us.Transaction.Rollback(us.Ctx)
}

// Commit фиксирует текущую транзакцию.
// Возвращает ошибку, если произошла ошибка при фиксации.
func (us *UsersStorage) Commit() error {
	return us.Transaction.Commit(us.Ctx)
}

// AddData запись данных в хранилище
func (us *UsersStorage) AddData(userID int, body []byte, dataType string) error {
	query := "INSERT INTO user_data (user_id, user_data, data_type) VALUES ($1, $2, $3) RETURNING id"

	userData := string(body)
	err := us.Conn.QueryRow(us.Ctx, query, userID, userData, dataType).Scan(&userID)
	if err != nil {
		return err
	}

	return nil
}
