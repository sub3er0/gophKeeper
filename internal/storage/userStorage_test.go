package storage

import (
	"errors"
	"testing"
)

// MockUserStorage представляет собой мок класса UserStorageInterface для тестирования.
type MockUserStorage struct {
	users     map[string]int
	cookies   map[string]int
	userData  map[int][]UserData
	userCount int
}

// NewMockUserStorage создает новый экземпляр MockUserStorage.
func NewMockUserStorage() *MockUserStorage {
	return &MockUserStorage{
		users:    make(map[string]int),
		cookies:  make(map[string]int),
		userData: make(map[int][]UserData),
	}
}

// Реализация методов интерфейса

func (m *MockUserStorage) IsCookieExist(uniqueID string) bool {
	_, exists := m.cookies[uniqueID]
	return exists
}

func (m *MockUserStorage) GetUserID(login string) (int, error) {
	if userID, exists := m.users[login]; exists {
		return userID, nil
	}
	return 0, errors.New("пользователь не найден")
}

func (m *MockUserStorage) SaveUserCookie(cookieValue string, userID int) error {
	m.cookies[cookieValue] = userID
	return nil
}

func (m *MockUserStorage) UpdateUserCookie(cookieValue string, userID int) error {
	if _, exists := m.cookies[cookieValue]; !exists {
		return errors.New("кука не существует")
	}
	m.cookies[cookieValue] = userID
	return nil
}

func (m *MockUserStorage) SaveUser(login string, password string) (int, error) {
	if _, exists := m.users[login]; exists {
		return 0, errors.New("пользователь уже существует")
	}
	m.userCount++
	m.users[login] = m.userCount
	return m.userCount, nil
}

func (m *MockUserStorage) CheckUserCredentials(login string, password string) (int, error) {
	if userID, exists := m.users[login]; exists {
		return userID, nil
	}
	return 0, errors.New("неверные учетные данные")
}

func (m *MockUserStorage) AddData(userID int, body []byte, dataType string) error {
	m.userData[userID] = append(m.userData[userID], UserData{
		ID:       uint(len(m.userData[userID]) + 1), // Инкрементальный ID
		UserID:   uint(userID),
		UserData: string(body),
		DataType: dataType,
	})
	return nil
}

func (m *MockUserStorage) GetData(userID int) ([]UserData, error) {
	if data, exists := m.userData[userID]; exists {
		return data, nil
	}
	return nil, errors.New("данные не найдены")
}

func (m *MockUserStorage) DeleteData(userID int, dataID int) error {
	if data, exists := m.userData[userID]; exists {
		for i, userData := range data {
			if userData.ID == uint(dataID) {
				m.userData[userID] = append(data[:i], data[i+1:]...) // Удаляем элемент
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

func (m *MockUserStorage) EditData(userID int, body []byte, dataType string, dataID int) error {
	if data, exists := m.userData[userID]; exists {
		for i, userData := range data {
			if userData.ID == uint(dataID) {
				userData.UserData = string(body) // Изменяем данные
				userData.DataType = dataType
				m.userData[userID][i] = userData // Обновляем значение в массиве
				return nil
			}
		}
		return errors.New("данные не найдены")
	}
	return errors.New("пользователь не найден")
}

func TestEncryptDecrypt(t *testing.T) {
	originalText := "Hello, GophKeeper!"
	encrypted, err := Encrypt(originalText)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if decrypted != originalText {
		t.Errorf("Expected decrypted text '%s', got '%s'", originalText, decrypted)
	}
}

func TestDecrypt_InvalidBase64String(t *testing.T) {
	_, err := Decrypt("invalid_base64")
	if err == nil {
		t.Error("Expected error for invalid base64 input, got none")
	}
}

func TestEncrypt_EmptyString(t *testing.T) {
	emptyString := ""
	encrypted, err := Encrypt(emptyString)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if decrypted != emptyString {
		t.Errorf("Expected decrypted text to be an empty string, got '%s'", decrypted)
	}
}
