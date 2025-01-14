package cookie

import (
	"net/http"
)

// AuthMiddleware оборачивает HTTP-обработчик для проверки аутентификации пользователя.
// Этот мидлвар проверяет наличие куки с именем user_info и ее валидность.
func (cm *CookieManager) AuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieName)
		if err != nil || !VerifyCookie(cookie.Value) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, _ := GetUserIDFromCookie(cookie.Value)
		_, err = cm.Storage.GetUserID(userID)

		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}
