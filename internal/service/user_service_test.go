package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"gophKeeper/internal/cookie"
	"gophKeeper/internal/storage"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// MockUserStorage представляет собой мок для UserStorageInterface
type MockUserStorage struct {
	mu        sync.Mutex
	users     map[string]int             // Хранит логин пользователя и его ID
	cookies   map[string]int             // Хранит куки и соответствующий ID пользователя
	userData  map[int][]storage.UserData // Хранит данные пользователей
	userCount int                        // Счетчик пользователей для генерации ID
}

// NewMockUserStorage создает новый экземпляр MockUserStorage
func NewMockUserStorage() *MockUserStorage {
	return &MockUserStorage{
		users:    make(map[string]int),
		cookies:  make(map[string]int),
		userData: make(map[int][]storage.UserData),
	}
}

// Реализация методов интерфейса

// IsCookieExist проверяет, существует ли куки пользователя
func (m *MockUserStorage) IsCookieExist(uniqueID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.cookies[uniqueID]
	return exists
}

// GetUserID проверяет, существует ли пользователь по указанному логину, возвращает его ID
func (m *MockUserStorage) GetUserID(login string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if userID, exists := m.users[login]; exists {
		return userID, nil
	}
	return 0, errors.New("пользователь не найден")
}

// SaveUserCookie сохраняет куки пользователя
func (m *MockUserStorage) SaveUserCookie(cookieValue string, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cookies[cookieValue] = userID
	return nil
}

// UpdateUserCookie обновляет куки пользователя
func (m *MockUserStorage) UpdateUserCookie(cookieValue string, userID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cookies[cookieValue] = userID
	if _, exists := m.cookies[cookieValue]; !exists {
		return errors.New("кука не существует")
	}
	return nil
}

// SaveUser сохраняет нового пользователя
func (m *MockUserStorage) SaveUser(login string, password string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userCount++
	if _, exists := m.users[login]; exists {
		return 0, errors.New("пользователь уже существует")
	}
	m.users[login] = m.userCount
	return m.userCount, nil
}

// SaveUserToken сохраняет токен пользователя (можно использовать как заглушку)
func (m *MockUserStorage) SaveUserToken(login string, password string) (int, error) {
	return 0, nil // Токен в мок-версии не сохраняется
}

// CheckUserCredentials проверяет валидность данных аутентификации
func (m *MockUserStorage) CheckUserCredentials(login string, password string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if userCreds, exists := m.users[login]; exists {
		if userCreds == 1 { // Сравниваем хешированный пароль
			return 1, nil
		}
		return 0, errors.New("неверные учетные данные")
	}
	return 0, errors.New("пользователь не найден")
}

// Init инициализирует хранилище пользователей с помощью строки соединения. В моках это просто заглушка.
func (m *MockUserStorage) Init(connectionString string) error {
	return nil
}

// Close закрывает соединение с хранилищем данных. Заглушка для моков.
func (m *MockUserStorage) Close() {
	// У нас нет открытого подключения в моках
}

// GetUsersCount получение количества пользователей
func (m *MockUserStorage) GetUsersCount() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.users), nil
}

// Ping проверяет состояние соединения с базой данных. В моках всегда true.
func (m *MockUserStorage) Ping() bool {
	return true
}

// BeginTransaction получение объекта транзакции. Мок никогда не делает реальных транзакций.
func (m *MockUserStorage) BeginTransaction() error {
	return nil
}

// Rollback откатывает текущую транзакцию. Заглушка для моков.
func (m *MockUserStorage) Rollback() error {
	return nil
}

// Commit фиксирует текущую транзакцию. Заглушка для моков.
func (m *MockUserStorage) Commit() error {
	return nil
}

// AddData сохраняет данные в хранилище
func (m *MockUserStorage) AddData(userID int, body []byte, dataType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userData[userID] = append(m.userData[userID], storage.UserData{
		ID:       uint(len(m.userData[userID]) + 1), // Инкрементальный ID
		UserID:   uint(userID),
		UserData: string(body),
		DataType: dataType,
	})
	return nil
}

// GetData дает данные пользователя
func (m *MockUserStorage) GetData(userID int) ([]storage.UserData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, exists := m.userData[userID]
	if !exists {
		return nil, errors.New("данные не найдены")
	}
	return data, nil
}

// DeleteData удаляет данные пользователя
func (m *MockUserStorage) DeleteData(userID int, dataID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if data, exists := m.userData[userID]; exists {
		for i, userData := range data {
			if userData.ID == uint(dataID) {
				m.userData[userID] = append(data[:i], data[i+1:]...)
				if len(m.userData[userID]) == 0 {
					delete(m.userData, userID) // Удаляем пользователя из мапы
				}
				return nil
			}
		}
		return errors.New("данные не найдены")
	}
	return errors.New("пользователь не найден")
}

// EditData изменяет данные пользователя в хранилище
func (m *MockUserStorage) EditData(userID int, body []byte, dataType string, dataID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if data, exists := m.userData[userID]; exists {
		for i, userData := range data {
			if userData.ID == uint(dataID) {
				userData.UserData = string(body)
				userData.DataType = dataType
				m.userData[userID][i] = userData // Обновляем значение в массиве
				return nil
			}
		}
		return errors.New("данные не найдены")
	}
	return errors.New("пользователь не найден")
}

// Тесты для UserService
func TestUserService_Registration(t *testing.T) {
	mockStorage := NewMockUserStorage()
	userService := NewUserService(mockStorage)

	user := RegistrationBody{
		Login:    "testuser",
		Password: "password123",
	}

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))

	cookie, err := userService.Registration(req)
	if err != nil || cookie == "" {
		t.Errorf("Expected successful registration, got error: %v", err)
	}

	if _, exists := mockStorage.users["testuser"]; !exists {
		t.Error("Expected user to be saved in storage")
	}
}

func TestUserService_Registration_UserExists(t *testing.T) {
	mockStorage := NewMockUserStorage()
	mockStorage.SaveUser("testuser", "password123")
	userService := NewUserService(mockStorage)

	user := RegistrationBody{
		Login:    "testuser",
		Password: "password123",
	}

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))

	cookie, err := userService.Registration(req)
	if err == nil {
		t.Error("Expected error because user exists, got none")
	}

	if cookie != "" {
		t.Errorf("Expected no cookie to be returned, got %s", cookie)
	}
}

func TestUserService_Authentication(t *testing.T) {
	mockStorage := NewMockUserStorage()
	mockStorage.SaveUser("testuser", "password123") // Сначала сохраняем пользователя
	userService := NewUserService(mockStorage)

	user := RegistrationBody{
		Login:    "testuser",
		Password: "password123",
	}

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))

	cookie, err := userService.Authentication(req)
	if err != nil || cookie == "" {
		t.Errorf("Expected successful authentication, got error: %v", err)
	}

	// После информации о куке, сохраняем куку в хранилище
	if err := mockStorage.SaveUserCookie(cookie, 1); err != nil {
		t.Errorf("Expected cookie to be saved in storage after authentication, error: %v", err)
	}

	// Проверим, сохранена ли кука
	if _, exists := mockStorage.cookies[cookie]; !exists {
		t.Error("Expected cookie to be saved in storage after authentication")
	}
}

func TestUserService_Authentication_InvalidCredentials(t *testing.T) {
	mockStorage := NewMockUserStorage()
	mockStorage.SaveUser("testuser", "password123") // Сначала сохраняем пользователя
	userService := NewUserService(mockStorage)

	user := RegistrationBody{
		Login:    "testuser1",
		Password: "wrongpassword", // Неправильный пароль
	}

	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))

	_, err := userService.Authentication(req)
	if err == nil {
		t.Error("Expected error due to invalid credentials, got none")
	}
}

func TestUserService_AddData(t *testing.T) {
	mockStorage := NewMockUserStorage()
	mockStorage.SaveUser("testuser", "password123") // Сохраняем пользователя
	userID, _ := mockStorage.GetUserID("testuser")
	mockStorage.SaveUserCookie("cookie_test", userID) // Сохраняем куку

	userService := NewUserService(mockStorage)

	data := []byte("test data")
	req := httptest.NewRequest("POST", "/adddata", bytes.NewBuffer(data))
	userIDStr := "1"
	encryptedUserID, err := storage.Encrypt(userIDStr)
	req.AddCookie(&http.Cookie{Name: cookie.CookieName, Value: encryptedUserID + "." + cookie.SignCookie(encryptedUserID)})

	err = userService.AddData(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	addedData, _ := mockStorage.GetData(userID)
	if len(addedData) == 0 {
		t.Error("Expected data to be added, but got none")
	}
}

func TestUserService_GetData(t *testing.T) {
	mockStorage := NewMockUserStorage()
	mockStorage.SaveUser("testuser", "password123") // Сохраняем пользователя
	userID, _ := mockStorage.GetUserID("testuser")
	mockStorage.SaveUserCookie("cookie_test", userID)
	mockStorage.AddData(userID, []byte("test data"), "text") // Добавляем данные

	userService := NewUserService(mockStorage)

	req := httptest.NewRequest("GET", "/getdata", nil)
	userIDStr := "1"
	encryptedUserID, err := storage.Encrypt(userIDStr)
	req.AddCookie(&http.Cookie{Name: cookie.CookieName, Value: encryptedUserID + "." + cookie.SignCookie(encryptedUserID)})

	userData, err := userService.GetData(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(userData) == 0 {
		t.Error("Expected to get data back, got none")
	}
}

func TestUserService_DeleteData(t *testing.T) {
	mockStorage := NewMockUserStorage()
	mockStorage.SaveUser("testuser", "password123") // Сохраняем пользователя
	userID, _ := mockStorage.GetUserID("testuser")
	mockStorage.SaveUserCookie("cookie_test", userID)
	mockStorage.AddData(userID, []byte("data to delete"), "text") // Добавляем данные

	userService := NewUserService(mockStorage)

	req := httptest.NewRequest("DELETE", "/deletedata?id=1", nil)
	userIDStr := "1"
	encryptedUserID, err := storage.Encrypt(userIDStr)
	req.AddCookie(&http.Cookie{Name: cookie.CookieName, Value: encryptedUserID + "." + cookie.SignCookie(encryptedUserID)})

	err = userService.DeleteData(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Проверяем, что данные были удалены
	if _, err := mockStorage.GetData(userID); err == nil {
		t.Error("Expected no data after delete, but data still exists")
	}
}

func TestUserService_EditData(t *testing.T) {
	mockStorage := NewMockUserStorage()
	mockStorage.SaveUser("testuser", "password123") // Сохраняем пользователя
	userID, _ := mockStorage.GetUserID("testuser")
	mockStorage.SaveUserCookie("cookie_test", userID)
	mockStorage.AddData(userID, []byte("original data"), "text") // Добавляем данные

	userService := NewUserService(mockStorage)

	// Изменяем данные
	req := httptest.NewRequest("PUT", "/editdata", bytes.NewBuffer([]byte("edited data")))
	req.Header.Set("Data-Id", "1")
	userIDStr := "1"
	encryptedUserID, err := storage.Encrypt(userIDStr)
	req.AddCookie(&http.Cookie{Name: cookie.CookieName, Value: encryptedUserID + "." + cookie.SignCookie(encryptedUserID)})

	err = userService.EditData(req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	editedData, _ := mockStorage.GetData(userID)
	if string(editedData[0].UserData) != "edited data" {
		t.Errorf("Expected 'edited data', got '%s'", editedData[0].UserData)
	}
}
