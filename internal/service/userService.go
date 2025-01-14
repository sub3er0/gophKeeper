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
		return "", nil // Пользователь уже существует
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
	encryptedUserID, err := cookie.Encrypt(userIDStr)
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
	encryptedUserID, err := cookie.Encrypt(userIDStr)
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
	userIDString, _ := cookie.GetUserIDFromCookie(userCookie.Value)
	userID, err := cookie.Decrypt(userIDString)
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
