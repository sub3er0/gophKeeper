package storage

// Users представляет структуру таблицы users.
type Users struct {
	ID        uint   `gorm:"primaryKey"`
	Login     string `gorm:"uniqueIndex;size:100"`
	Password  string `gorm:"size:256"`
	IsDeleted bool   `gorm:"default:false"`
}

// UserCookie представляет структуру таблицы users_cookie.
type UserCookie struct {
	ID          uint   `gorm:"primaryKey"`
	UserID      uint   `gorm:"uniqueIndex"`
	CookieValue string `gorm:"size:100"`
}

// UserData представляет структуру таблицы user_data.
type UserData struct {
	ID       uint `gorm:"primaryKey"`
	UserID   uint
	UserData string `gorm:"size:256"`
	DataType string `gorm:"size:100"`
}
