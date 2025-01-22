package cookie

import (
	"errors"
	"gophKeeper/internal/storage"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// MockUserStorage представляет мок для UserStorageInterface
type MockUserStorage struct {
	users     map[string]int             // хранит пары логин -> ID пользователя
	cookies   map[string]int             // хранит куки -> ID пользователя
	userData  map[int][]storage.UserData // хранит данные пользователей по userID
	userCount int                        // симуляция количества пользователей
	mu        sync.Mutex                 // защита от конкурентного доступа
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
	// В Mock всегда возвращаем true, т.к. никаких соединений не проводим
	return true
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

// TestSignCookie проверяет, правильно ли работает подписка куки.
func TestSignCookie(t *testing.T) {
	data := "test_data"
	signature := SignCookie(data)

	if signature == "" {
		t.Errorf("expected non-empty signature")
	}

	// Проверяем, что подпись генерируется каждый раз по одному и тому же входу
	expected := SignCookie(data)
	if signature != expected {
		t.Errorf("expected signature '%s', got '%s'", expected, signature)
	}
}

// TestVerifyCookie проверяет, правильно ли работает верификация куки.
func TestVerifyCookie(t *testing.T) {
	data := "user_id"
	signedData := data + "." + SignCookie(data)

	if !VerifyCookie(signedData) {
		t.Errorf("expected true for valid cookie, got false")
	}

	if VerifyCookie("invalid.cookie") {
		t.Errorf("expected false for invalid cookie, got true")
	}
}

// TestGetUserIDFromCookie проверяет извлечение идентификатора пользователя из куки.
func TestGetUserIDFromCookie(t *testing.T) {
	userID := "user_id"
	cookie := userID + "." + SignCookie(userID)

	parsedID, valid := GetUserIDFromCookie(cookie)
	if !valid || parsedID != userID {
		t.Errorf("expected userID '%s', got '%s'", userID, parsedID)
	}

	_, valid = GetUserIDFromCookie("invalid.cookie.asdf")
	if valid {
		t.Errorf("expected false for invalid cookie, got true")
	}
}

// TestCookieHandler проверяет работу CookieHandler.
func TestCookieHandler(t *testing.T) {
	mockStorage := &MockUserStorage{}
	cm := &CookieManager{Storage: mockStorage}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	})

	cookieValue := "user_id." + SignCookie("user_id")
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: CookieName, Value: cookieValue})

	// Создаем реквизит-ответчик для записи ответа
	rr := httptest.NewRecorder()

	cm.CookieHandler(handler).ServeHTTP(rr, req)

	// Проверяем статус-код и тело ответа
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := "Success"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}

	// Тестируем случай, когда куки нет
	req = httptest.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	cm.CookieHandler(handler).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}
