package user

import "time"

type User struct {
	ID          string    `json:"id" gorm:"primaryKey;type:UUID;default:gen_random_uuid()"`
	Email       string    `json:"email" gorm:"unique;not null"`
	Password    string    `json:"-" gorm:"not null"`
	UserProfile Profile   `json:"user_profile"`
	Role        string    `json:"role" gorm:"not null;default:'user'"`
	Status      string    `json:"status" gorm:"not null;default:'active'"`
	LastLogin   time.Time `json:"last_login"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Profile struct {
	ID          string    `json:"id" gorm:"primaryKey;type:UUID;default:gen_random_uuid()"`
	UserID      string    `json:"user_id" gorm:"unique;not null"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Bio         string    `json:"bio"`
	Language    string    `json:"language" gorm:"not null;default:'en'"`
	Timezone    string    `json:"timezone" gorm:"not null;default:'UTC'"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
