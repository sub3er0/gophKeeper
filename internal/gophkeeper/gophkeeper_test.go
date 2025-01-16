package gophkeeper

import (
	"bytes"
	"encoding/json"
	"errors"
	"gophKeeper/internal/cookie"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"gophKeeper/internal/storage"
)

// MockUserService представляет собой мок класса UserServiceInterface для тестирования.
type MockUserService struct {
	RegistrationFunc   func(r *http.Request) (string, error)
	AuthenticationFunc func(r *http.Request) (string, error)
	AddDataFunc        func(r *http.Request) error
	GetDataFunc        func(r *http.Request) ([]storage.UserData, error)
	DeleteDataFunc     func(r *http.Request) error
	EditDataFunc       func(r *http.Request) error
}

// MockUserStorage представляет мок для UserStorageInterface
type MockUserStorage struct {
	users     map[string]int             // хранит пары логин -> ID пользователя
	cookies   map[string]int             // хранит куки -> ID пользователя
	userData  map[int][]storage.UserData // хранит данные пользователей по userID
	userCount int                        // симуляция количества пользователей
	mu        sync.Mutex                 // защита от конкурентного доступа
	PingFunc  func() bool
}

// NewMockUserStorage создает новый экземпляр MockUserStorage
func NewMockUserStorage() *MockUserStorage {
	return &MockUserStorage{
		users:    make(map[string]int),
		cookies:  make(map[string]int),
		userData: make(map[int][]storage.UserData),
	}
}

// IsCookieExist проверяет, существует ли куки пользователя
func (m *MockUserStorage) IsCookieExist(uniqueID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.cookies[uniqueID]
	return exists
}

// GetUserID проверяет, существует ли пользователь по указанному уникальному идентификатору.
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
	if _, exists := m.cookies[cookieValue]; !exists {
		return errors.New("кука не существует")
	}
	m.cookies[cookieValue] = userID
	return nil
}

// SaveUser сохраняет пользователя
func (m *MockUserStorage) SaveUser(login string, password string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.users[login]; exists {
		return 0, errors.New("пользователь уже существует")
	}
	m.userCount++
	m.users[login] = m.userCount
	return m.userCount, nil
}

// SaveUserToken сохраняет токен пользователя
func (m *MockUserStorage) SaveUserToken(login string, password string) (int, error) {
	// В Mock реализации это может быть простой сохранение пользователя или фокус на другой логике
	return 0, nil
}

// CheckUserCredentials проверяет валидность данных аутентификации
func (m *MockUserStorage) CheckUserCredentials(login string, password string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if userID, exists := m.users[login]; exists {
		// Здесь может быть логика проверки пароля, для простоты игнорируем пароль
		return userID, nil
	}
	return 0, errors.New("неверные учетные данные")
}

// Init инициализирует хранилище пользователей с помощью строки соединения.
func (m *MockUserStorage) Init(connectionString string) error {
	// В Mock реализации можно оставить пустым
	return nil
}

// Close закрывает соединение с хранилищем данных.
func (m *MockUserStorage) Close() {
	// В Mock можно оставить пустым
}

// GetUsersCount получение количества пользователей
func (m *MockUserStorage) GetUsersCount() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.users), nil
}

// Ping проверяет состояние соединения с базой данных.
func (m *MockUserStorage) Ping() bool {
	return m.PingFunc()
}

// BeginTransaction получение объекта транзакции
func (m *MockUserStorage) BeginTransaction() error {
	// В Mock можно оставить пустым
	return nil
}

// Rollback откатывает текущую транзакцию.
func (m *MockUserStorage) Rollback() error {
	// В Mock можно оставить пустым
	return nil
}

// Commit фиксирует текущую транзакцию.
func (m *MockUserStorage) Commit() error {
	// В Mock можно оставить пустым
	return nil
}

// AddData запись данных в хранилище
func (m *MockUserStorage) AddData(userID int, body []byte, dataType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userData[userID] = append(m.userData[userID], storage.UserData{
		ID:       uint(len(m.userData[userID]) + 1), // Простой автоинкрементный ID
		UserID:   uint(userID),
		UserData: string(body),
		DataType: dataType,
	})
	return nil
}

// GetData получить данные пользователя
func (m *MockUserStorage) GetData(userID int) ([]storage.UserData, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, exists := m.userData[userID]
	if !exists {
		return nil, errors.New("данные не найдены")
	}
	return data, nil
}

// DeleteData удаление данных пользователя
func (m *MockUserStorage) DeleteData(userID int, dataID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if data, exists := m.userData[userID]; exists {
		for i, userData := range data {
			if userData.ID == uint(dataID) {
				m.userData[userID] = append(data[:i], data[i+1:]...) // Удалить элемент
				return nil
			}
		}
		return errors.New("данные не найдены")
	}
	return errors.New("пользователь не найден")
}

// EditData изменение данных в хранилище
func (m *MockUserStorage) EditData(userID int, body []byte, dataType string, dataID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if data, exists := m.userData[userID]; exists {
		for i, userData := range data {
			if userData.ID == uint(dataID) {
				userData.UserData = string(body)
				userData.DataType = dataType
				m.userData[userID][i] = userData // Обновление значения в массиве
				return nil
			}
		}
		return errors.New("данные не найдены")
	}
	return errors.New("пользователь не найден")
}

func (m *MockUserService) Registration(r *http.Request) (string, error) {
	return m.RegistrationFunc(r)
}

func (m *MockUserService) Authentication(r *http.Request) (string, error) {
	return m.AuthenticationFunc(r)
}

func (m *MockUserService) AddData(r *http.Request) error {
	return m.AddDataFunc(r)
}

func (m *MockUserService) GetData(r *http.Request) ([]storage.UserData, error) {
	return m.GetDataFunc(r)
}

func (m *MockUserService) DeleteData(r *http.Request) error {
	return m.DeleteDataFunc(r)
}

func (m *MockUserService) EditData(r *http.Request) error {
	return m.EditDataFunc(r)
}

func TestPingHandler(t *testing.T) {
	// Создаем мок для UserStorage, который возвращает false при проверке соединения
	mockStorage := &MockUserStorage{
		PingFunc: func() bool {
			return true
		},
	}

	keeper := &GophKeeper{
		Storage: mockStorage,
	}

	req := httptest.NewRequest("GET", "/ping", nil)
	rec := httptest.NewRecorder()

	keeper.PingHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", res.StatusCode)
	}
}

func TestPingHandler_Error(t *testing.T) {
	// Создаем мок для UserStorage, который возвращает false при проверке соединения
	mockStorage := &MockUserStorage{
		PingFunc: func() bool {
			return false
		},
	}

	keeper := &GophKeeper{
		Storage: mockStorage,
	}

	req := httptest.NewRequest("GET", "/ping", nil)
	rec := httptest.NewRecorder()

	keeper.PingHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status InternalServerError, got %v", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	expectedMessage := "Connection error\n"
	if string(bodyBytes) != expectedMessage {
		t.Errorf("Expected response body '%s', got '%s'", expectedMessage, string(bodyBytes))
	}
}

func TestRegistrationHandler(t *testing.T) {
	mockUserService := &MockUserService{
		RegistrationFunc: func(r *http.Request) (string, error) {
			return "test_cookie_value", nil // Симулируем успешную регистрацию
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	user := map[string]string{"login": "testuser", "password": "password123"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	keeper.RegistrationHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", res.StatusCode)
	}

	cookieValue := rec.Header().Get("Set-Cookie")
	if cookieValue == "" {
		t.Errorf("Expected cookie to be set, got none")
	}
}

func TestRegistrationHandler_Error(t *testing.T) {
	mockUserService := &MockUserService{
		RegistrationFunc: func(r *http.Request) (string, error) {
			return "", errors.New("ошибка при регистрации") // Симулируем ошибку
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	user := map[string]string{"login": "testuser", "password": "password123"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	keeper.RegistrationHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, got %v", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	expectedErrorMessage := "ошибка при регистрации\n"
	if string(bodyBytes) != expectedErrorMessage {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMessage, string(bodyBytes))
	}
}

func TestAuthenticationHandler(t *testing.T) {
	mockUserService := &MockUserService{
		AuthenticationFunc: func(r *http.Request) (string, error) {
			return "test_cookie_value", nil // Симулируем успешную аутентификацию
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	user := map[string]string{"login": "testuser", "password": "password123"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	keeper.AuthenticationHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", res.StatusCode)
	}

	cookieValue := rec.Header().Get("Set-Cookie")
	if cookieValue == "" {
		t.Errorf("Expected cookie to be set, got none")
	}
}

func TestAuthenticationHandler_Error(t *testing.T) {
	mockUserService := &MockUserService{
		AuthenticationFunc: func(r *http.Request) (string, error) {
			return "", errors.New("неверный логин или пароль") // Симулируем ошибку аутентификации
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	user := map[string]string{"login": "testuser", "password": "wrongpassword"}
	body, _ := json.Marshal(user)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
	rec := httptest.NewRecorder()

	keeper.AuthenticationHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, got %v", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	expectedErrorMessage := "неверный логин или пароль\n"
	if string(bodyBytes) != expectedErrorMessage {
		t.Errorf("Expected error message '%s', got '%s'", expectedErrorMessage, string(bodyBytes))
	}
}

func TestAddDataHandler(t *testing.T) {
	mockUserService := &MockUserService{
		AddDataFunc: func(r *http.Request) error {
			return nil // Симулируем успешное добавление данных
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	data := map[string]string{"data": "test data"}
	dataBody, _ := json.Marshal(data)
	req := httptest.NewRequest("POST", "/adddata", bytes.NewBuffer(dataBody))
	rec := httptest.NewRecorder()

	keeper.AddDataHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", res.StatusCode)
	}
}

func TestAddDataHandler_Error(t *testing.T) {
	mockUserService := &MockUserService{
		AddDataFunc: func(r *http.Request) error {
			return errors.New("ошибка при добавлении данных") // Симулируем ошибку
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	data := map[string]string{"data": "test data"}
	dataBody, _ := json.Marshal(data)
	req := httptest.NewRequest("POST", "/adddata", bytes.NewBuffer(dataBody))
	req.AddCookie(&http.Cookie{Name: cookie.CookieName, Value: "cookie_test"}) // Добавляем куку для авторизации
	rec := httptest.NewRecorder()

	keeper.AddDataHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status InternalServerError, got %v", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	expectedErrorMsg := "ошибка при добавлении данных\n"
	if string(bodyBytes) != expectedErrorMsg {
		t.Errorf("Expected error message in response, got '%s'", string(bodyBytes))
	}
}

func TestGetDataHandler(t *testing.T) {
	mockUserService := &MockUserService{
		GetDataFunc: func(r *http.Request) ([]storage.UserData, error) {
			return []storage.UserData{
				{ID: 1, UserID: 1, UserData: "test data", DataType: "text"},
			}, nil // Симулируем успешное получение данных
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	req := httptest.NewRequest("GET", "/getdata", nil)
	rec := httptest.NewRecorder()

	keeper.GetDataHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", res.StatusCode)
	}

	var userDataArray []storage.UserData
	if err := json.NewDecoder(rec.Body).Decode(&userDataArray); err != nil {
		t.Errorf("Error decoding response body: %v", err)
	}

	if len(userDataArray) != 1 || userDataArray[0].UserData != "test data" {
		t.Errorf("Expected to receive test data, got %v", userDataArray)
	}
}

func TestGetDataHandler_NoCookie(t *testing.T) {
	mockUserService := &MockUserService{
		GetDataFunc: func(r *http.Request) ([]storage.UserData, error) {
			return nil, errors.New("кука отсутствует") // Симуляция отсутствующей куки
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	req := httptest.NewRequest("GET", "/getdata", nil) // Не добавляем куку
	rec := httptest.NewRecorder()

	keeper.GetDataHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status InternalServerError, got %v", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	expectedMessage := "Произошла ошибка при получении данных\nnull\n"
	if string(bodyBytes) != expectedMessage {
		t.Errorf("Expected error message to be '%s', got '%s'", expectedMessage, string(bodyBytes))
	}
}

func TestGetDataHandler_DataNotFound(t *testing.T) {
	mockUserService := &MockUserService{
		GetDataFunc: func(r *http.Request) ([]storage.UserData, error) {
			return nil, errors.New("данные не найдены") // Симуляция отсутствия данных
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	req := httptest.NewRequest("GET", "/getdata", nil)
	req.AddCookie(&http.Cookie{Name: cookie.CookieName, Value: "cookie_test"}) // Устанавливаем куку
	rec := httptest.NewRecorder()

	keeper.GetDataHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status InternalServerError, got %v", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	expectedMessage := "Произошла ошибка при получении данных\nnull\n"
	if string(bodyBytes) != expectedMessage {
		t.Errorf("Expected error message to be '%s', got '%s'", expectedMessage, string(bodyBytes))
	}
}

func TestDeleteDataHandler(t *testing.T) {
	mockUserService := &MockUserService{
		DeleteDataFunc: func(r *http.Request) error {
			return nil // Симулируем успешное удаление данных
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	req := httptest.NewRequest("DELETE", "/deletedata", nil)
	rec := httptest.NewRecorder()

	keeper.DeleteDataHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", res.StatusCode)
	}
}

func TestDeleteDataHandler_Error(t *testing.T) {
	mockUserService := &MockUserService{
		DeleteDataFunc: func(r *http.Request) error {
			return errors.New("Произошла ошибка при получении данных") // Симулируем ошибку
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	req := httptest.NewRequest("DELETE", "/deletedata?id=1", nil) // Указываем ID данных для удаления
	rec := httptest.NewRecorder()

	keeper.DeleteDataHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status InternalServerError, got %v", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}
	expectedErrorMsg := "Произошла ошибка при получении данных\n"
	if string(bodyBytes) != expectedErrorMsg {
		t.Errorf("Expected error message in response, got '%s'", bodyBytes)
	}
}

func TestEditDataHandler(t *testing.T) {
	mockUserService := &MockUserService{
		EditDataFunc: func(r *http.Request) error {
			return nil // Симулируем успешное редактирование данных
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	data := map[string]string{"data": "updated data"}
	dataBody, _ := json.Marshal(data)
	req := httptest.NewRequest("PUT", "/editdata", bytes.NewBuffer(dataBody))
	rec := httptest.NewRecorder()

	keeper.EditDataHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", res.StatusCode)
	}
}

func TestEditDataHandler_Error(t *testing.T) {
	mockUserService := &MockUserService{
		EditDataFunc: func(r *http.Request) error {
			return errors.New("ошибка обновления данных") // Возвращаем ошибку
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	data := map[string]string{"data": "updated data"}
	dataBody, _ := json.Marshal(data)
	req := httptest.NewRequest("PUT", "/editdata", bytes.NewBuffer(dataBody))
	req.Header.Set("Data-Id", "1")                                             // Указываем ID данных для редактирования
	req.AddCookie(&http.Cookie{Name: cookie.CookieName, Value: "cookie_test"}) // Добавляем куку
	rec := httptest.NewRecorder()

	keeper.EditDataHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status InternalServerError, got %v", res.StatusCode)
	}

	bodyBytes, _ := io.ReadAll(rec.Body)
	expectedErrorMsg := "ошибка обновления данных\n"
	if string(bodyBytes) != expectedErrorMsg {
		t.Errorf("Expected error message in response, got %s", string(bodyBytes))
	}
}

func TestEditDataHandler_InvalidDataID(t *testing.T) {
	mockUserService := &MockUserService{
		EditDataFunc: func(r *http.Request) error {
			return errors.New("данные не найдены") // Симулируем ситуацию, когда данные не найдены
		},
	}

	keeper := &GophKeeper{
		UserService: mockUserService,
	}

	data := map[string]string{"data": "updated data"}
	dataBody, _ := json.Marshal(data)
	req := httptest.NewRequest("PUT", "/editdata", bytes.NewBuffer(dataBody))
	req.Header.Set("Data-Id", "999")                                           // Неверный ID данных для редактирования
	req.AddCookie(&http.Cookie{Name: cookie.CookieName, Value: "cookie_test"}) // Добавляем куку
	rec := httptest.NewRecorder()

	keeper.EditDataHandler(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status InternalServerError, got %v", res.StatusCode)
	}

	bodyBytes, _ := io.ReadAll(rec.Body)
	expectedErrorMsg := "данные не найдены\n"
	if string(bodyBytes) != expectedErrorMsg {
		t.Errorf("Expected error message in response, got %s", string(bodyBytes))
	}
}
