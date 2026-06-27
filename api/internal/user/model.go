package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` //không bao giờ xuất hiện trong response JSON trả về client, dù có vô tình json.Encode(user)
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
