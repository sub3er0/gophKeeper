package cookie

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"gophKeeper/internal/storage"
	"net/http"
	"strings"
)

// CookieManager управляет аутентификацией и обработкой куки в приложении.
// Он предоставляет методы для установки, проверки и получения значений куки.
type CookieManager struct {
	// Storage используется для взаимодействия с хранилищем пользователей.
	Storage storage.UserStorageInterface

	// ActualCookieValue содержит текущее значение куки аутентификации пользователя.
	ActualCookieValue string
}

// CookieManagerInterface определяет методы для работы с куками в приложении.
// Этот интерфейс позволяет управлять куками и реализовывать middleware для аутентификации.
type CookieManagerInterface interface {
	// CookieHandler оборачивает HTTP-обработчик для управления кукми.
	// Возвращает обработчик, который изменён для работы с куками.
	CookieHandler(h http.Handler) http.Handler

	// GetActualCookieValue возвращает значение актуальной куки для текущего пользователя.
	GetActualCookieValue() string
}

// GetActualCookieValue возвращает значение актуальной куки для текущего пользователя.
// Возвращает строку, представляющую актуальное значение куки.
func (cm *CookieManager) GetActualCookieValue() string {
	return cm.ActualCookieValue
}

var (
	SecretKey  = []byte("secret_key")
	CookieName = "user_info"
)

// SignCookie вычисляет HMAC-подпись для данных.
// Используется для проверки подлинности куки.
func SignCookie(data string) string {
	h := hmac.New(sha256.New, SecretKey)
	h.Write([]byte(data))
	signature := h.Sum(nil)

	return base64.RawURLEncoding.EncodeToString(signature)
}

// VerifyCookie проверяет корректность куки, сравнивая подпись с ожидаемой.
func VerifyCookie(str string) bool {
	parts := strings.Split(str, ".")
	if len(parts) != 2 {
		return false
	}

	dataStr := parts[0]
	sigStr := parts[1]
	expected := SignCookie(dataStr)

	return sigStr == expected
}

// GetUserIDFromCookie извлекает идентификатор пользователя из куки.
func GetUserIDFromCookie(str string) (string, bool) {
	parts := strings.Split(str, ".")
	if len(parts) != 2 {
		return "", false
	}

	return parts[0], true
}

// GenerateUserCookieID генерирует уникальный идентификатор пользователя.
func GenerateUserCookieID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// CookieHandler оборачивает HTTP-обработчик, добавляя логику работы с куками.
func (cm *CookieManager) CookieHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieName)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err != nil || !VerifyCookie(cookie.Value) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}
