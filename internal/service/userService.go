package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"gophKeeper/internal/cookie"
	"gophKeeper/internal/storage" // Замените на ваш актуальный путь
	"io"
	"log"
	"net/http"
	"strconv"
)

type RegistrationBody struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserService struct {
	Storage storage.UserStorageInterface
}

type UserServiceInterface interface {
	Registration(r *http.Request) (string, error)
	Authentication(r *http.Request) (string, error)
	AddData(r *http.Request) error
	GetData(r *http.Request) ([]storage.UserData, error)
	DeleteData(r *http.Request) error
	EditData(r *http.Request) error
}

// NewUserService Конструктор для создания экземпляра UserService
func NewUserService(storage storage.UserStorageInterface) *UserService {
	return &UserService{
		Storage: storage,
	}
}

// Registration регистрирует нового пользователя
func (us *UserService) Registration(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	var requestBody RegistrationBody
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		return "", err
	}

	_, err = us.Storage.GetUserID(requestBody.Login)
	if err == nil {
		return "", errors.New("пользователь существует")
	}

	err = us.Storage.BeginTransaction()
	if err != nil {
		return "", errors.New("ошибка регистрации пользователя")
	}

	defer func() {
		if err != nil {
			rollbackErr := us.Storage.Rollback()
			if rollbackErr != nil {
				log.Printf("Ошибка отката транзакции: %v", rollbackErr)
			}
		}
	}()

	userID, err := us.Storage.SaveUser(requestBody.Login, requestBody.Password)
	if err != nil {
		return "", err
	}

	userIDStr := fmt.Sprintf("%d", userID)
	encryptedUserID, err := storage.Encrypt(userIDStr)
	if err != nil {
		return "", err
	}

	newCookieValue := encryptedUserID + "." + cookie.SignCookie(encryptedUserID)
	err = us.Storage.SaveUserCookie(newCookieValue, userID)

	if err != nil {
		_ = us.Storage.Rollback()
		return "", err
	}

	if err = us.Storage.Commit(); err != nil {
		return "", err
	}

	return newCookieValue, nil
}

// Authentication авторизация пользователя
func (us *UserService) Authentication(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", err
	}

	var requestBody RegistrationBody
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		return "", err
	}

	err = us.Storage.BeginTransaction()
	if err != nil {
		return "", errors.New("ошибка аутентификации пользователя")
	}

	defer func() {
		if err != nil {
			rollbackErr := us.Storage.Rollback()
			if rollbackErr != nil {
				log.Printf("Ошибка отката транзакции: %v", rollbackErr)
			}
		}
	}()

	userID, err := us.Storage.CheckUserCredentials(requestBody.Login, requestBody.Password)
	if err != nil {
		return "", err
	}

	userIDStr := fmt.Sprintf("%d", userID)
	encryptedUserID, err := storage.Encrypt(userIDStr)
	if err != nil {
		return "", err
	}

	newCookieValue := encryptedUserID + "." + cookie.SignCookie(encryptedUserID)
	err = us.Storage.UpdateUserCookie(newCookieValue, userID)

	if err != nil {
		_ = us.Storage.Rollback()
		return "", err
	}

	if err = us.Storage.Commit(); err != nil {
		return "", err
	}

	return newCookieValue, nil
}

// AddData добавляет данные в хранилище
func (us *UserService) AddData(r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	userCookie, err := r.Cookie(cookie.CookieName)
	if err != nil {
		return err
	}
	userIDString, _ := cookie.GetUserIDFromCookie(userCookie.Value)
	userID, err := storage.Decrypt(userIDString)
	if err != nil {
		return err
	}
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return errors.New("userID must be an integer")
	}

	dataType := r.Header.Get("Data-Type")

	err = us.Storage.AddData(userIDInt, body, dataType)

	if err != nil {
		log.Printf("Ошибка при сохранении данных: %v", err)
		return errors.New("произошла ошибка при сохранении данных")
	}

	return nil
}

// GetData возвращает данные из хранилищ
func (us *UserService) GetData(r *http.Request) ([]storage.UserData, error) {
	userCookie, err := r.Cookie(cookie.CookieName)
	if err != nil {
		return nil, err
	}
	userIDString, _ := cookie.GetUserIDFromCookie(userCookie.Value)
	userID, err := storage.Decrypt(userIDString)
	if err != nil {
		return nil, err
	}
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return nil, errors.New("userID must be an integer")
	}

	UserDataArray, err := us.Storage.GetData(userIDInt)

	if err != nil {
		log.Printf("Ошибка при получении данных: %v", err)
		return nil, errors.New("произошла ошибка при получении данных")
	}

	return UserDataArray, nil
}

// DeleteData удаление данных
func (us *UserService) DeleteData(r *http.Request) error {
	userCookie, err := r.Cookie(cookie.CookieName)
	if err != nil {
		return err
	}
	userIDString, _ := cookie.GetUserIDFromCookie(userCookie.Value)
	userID, err := storage.Decrypt(userIDString)
	if err != nil {
		return err
	}
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return errors.New("userID must be an integer")
	}

	queryParams := r.URL.Query()
	dataID, err := strconv.Atoi(queryParams.Get("id"))
	if err != nil {
		return errors.New("dataID must be an integer")
	}
	err = us.Storage.DeleteData(userIDInt, dataID)

	if err != nil {
		log.Printf("Ошибка при удалении данных: %v", err)
		return errors.New("произошла ошибка при удалении данных")
	}

	return nil
}

// EditData изменяет данные в хранилище
func (us *UserService) EditData(r *http.Request) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	userCookie, err := r.Cookie(cookie.CookieName)
	if err != nil {
		return err
	}
	userIDString, _ := cookie.GetUserIDFromCookie(userCookie.Value)
	userID, err := storage.Decrypt(userIDString)
	if err != nil {
		return err
	}
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		return errors.New("userID must be an integer")
	}

	dataID := r.Header.Get("Data-Id")
	dataIDInt, err := strconv.Atoi(dataID)
	if err != nil {
		return errors.New("dataID must be an integer")
	}
	dataType := r.Header.Get("Data-Type")

	err = us.Storage.EditData(userIDInt, body, dataType, dataIDInt)

	if err != nil {
		log.Printf("Ошибка при измнении данных: %v", err)
		return errors.New("произошла ошибка при измнении данных")
	}

	return nil
}
