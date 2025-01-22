package gophkeeper

import (
	"encoding/json"
	"gophKeeper/internal/cookie"
	"gophKeeper/internal/service"
	"gophKeeper/internal/storage"
	"log"
	"net/http"
	"time"
)

// GophKeeper представляет структуру, ответственную за обработку запросов
type GophKeeper struct {
	Storage storage.UserStorageInterface

	// ServerAddress определяет адрес HTTP-сервера, на котором будет работать приложение.
	ServerAddress string

	// BaseURL представляет базовый адрес, который используется для сокращённых URL.
	BaseURL string

	// CookieManager управляет аутентификацией и обработкой куки в приложении.
	CookieManager cookie.CookieManagerInterface

	// UserService сервис для бизнес лоигки
	UserService service.UserServiceInterface
}

// GophKeeperInterface - интерфейс для работы с GophKeeper
type GophKeeperInterface interface {
	// PingHandler Проверяет состояние соединения с репозиторием
	PingHandler(w http.ResponseWriter, r *http.Request)

	RegistrationHandler(w http.ResponseWriter, r *http.Request)
}

// PingHandler Проверяет состояние соединения с репозиторием
func (us *GophKeeper) PingHandler(w http.ResponseWriter, r *http.Request) {
	ok := us.Storage.Ping()

	if ok {
		w.WriteHeader(http.StatusOK)
		return
	} else {
		http.Error(w, "Connection error", http.StatusInternalServerError)
	}
}

// RegistrationHandler Регистрация нового пользователя
func (us *GophKeeper) RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	newCookieValue, err := us.UserService.Registration(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Ошибка при создании нового пользователя: %v", err)
		return
	}

	if newCookieValue == "" {
		http.Error(w, "Пользователь с таким логином уже существует", http.StatusConflict)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookie.CookieName,
		Value:    newCookieValue,
		Path:     "/",
		Expires:  time.Now().AddDate(10, 0, 0),
		HttpOnly: true,
		Secure:   false,
	})

	w.Write([]byte("Регистрация успешна!"))
}

// AuthenticationHandler Регистрация нового пользователя
func (us *GophKeeper) AuthenticationHandler(w http.ResponseWriter, r *http.Request) {
	cookieValue, err := us.UserService.Authentication(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Ошибка при авторизации, неверный логин или пароль: %v", err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookie.CookieName,
		Value:    cookieValue,
		Path:     "/",
		Expires:  time.Now().AddDate(10, 0, 0),
		HttpOnly: true,
		Secure:   false,
	})

	w.Write([]byte("Аутентификация успешна!"))
}

// AddDataHandler Добавление данных
func (us *GophKeeper) AddDataHandler(w http.ResponseWriter, r *http.Request) {
	err := us.UserService.AddData(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Ошибка при сохранении данных: %v", err)
		return
	}

	_, err = w.Write([]byte("Данные успешно сохранены!"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Ошибка при сохранении данных: %v", err)
		return
	}
}

// GetDataHandler получение данных пользователя
func (us *GophKeeper) GetDataHandler(w http.ResponseWriter, r *http.Request) {
	UserDataArray, err := us.UserService.GetData(r)
	if err != nil {
		log.Printf("Ошибка при получении данных: %v", err)
		http.Error(w, "Произошла ошибка при получении данных", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(UserDataArray); err != nil {
		log.Printf("Ошибка при кодировании данных в JSON: %v", err)
		http.Error(w, "Ошибка при кодировании данных", http.StatusInternalServerError)
	}
}

// DeleteDataHandler удаление данных пользователя
func (us *GophKeeper) DeleteDataHandler(w http.ResponseWriter, r *http.Request) {
	err := us.UserService.DeleteData(r)
	if err != nil {
		log.Printf("Ошибка при получении данных: %v", err)
		http.Error(w, "Произошла ошибка при получении данных", http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte("Данные успешно удалены!"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Внутренняя ошибка сервера %v", err)
		return
	}
}

// EditDataHandler изменение данных
func (us *GophKeeper) EditDataHandler(w http.ResponseWriter, r *http.Request) {
	err := us.UserService.EditData(r)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Ошибка при сохранении данных: %v", err)
		return
	}

	_, err = w.Write([]byte("Данные успешно изменены!"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Ошибка при изменении данных: %v", err)
		return
	}
}
