package user

import (
	"context"
	"strings"

	"github.com/deepwrite/serivces/gateway/pkg/database"
	"github.com/deepwrite/serivces/gateway/pkg/hash"
	"gorm.io/gorm"
)

func (u *User) IsActive() bool {
	return u.Status == "active"
}

func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

func (u *User) IsBanned() bool {
	return u.Status == "banned"
}

func (u *User) ChangePassword(ctx context.Context, newPassword string) error {
	hashPwd, err := hash.HashPassword(newPassword)
	if err != nil {
		return err
	}
	u.Password = hashPwd
	_, err = gorm.G[User](database.DB).
		Where("id = ?", u.ID).
		Updates(ctx, User{Password: u.Password})
	return err
}

func (u *User) VerifyPassword(password string) bool {
	return hash.VerifyPassword(u.Password, password)
}

func (u *User) ChangeEmail(ctx context.Context, newEmail string) error {
	email := strings.ToLower(strings.TrimSpace(newEmail))
	if email == "" {
		return gorm.ErrInvalidValue
	}

	u.Email = email
	_, err := gorm.G[User](database.DB).
		Where("id = ?", u.ID).
		Updates(ctx, User{Email: u.Email})
	return err
}

func GetUserByEmail(ctx context.Context, email string) (User, error) {
	return gorm.G[User](database.DB).Where("email = ?", email).First(ctx)
}

func Create(ctx context.Context, user *User) error {
	hashPwd, err := hash.HashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashPwd
	return gorm.G[User](database.DB).Create(ctx, user)
}

func GetUserByID(ctx context.Context, id string) (User, error) {
	return gorm.G[User](database.DB).Preload("UserProfile", nil).Where("id = ?", id).First(ctx)
}

func UpdateUserProfile(ctx context.Context, user User) error {
	profile := Profile{
		DisplayName: strings.TrimSpace(user.UserProfile.DisplayName),
		AvatarURL:   strings.TrimSpace(user.UserProfile.AvatarURL),
		Bio:         strings.TrimSpace(user.UserProfile.Bio),
		Language:    strings.TrimSpace(user.UserProfile.Language),
		Timezone:    strings.TrimSpace(user.UserProfile.Timezone),
	}

	affected, err := gorm.G[Profile](database.DB).
		Where("user_id = ?", user.ID).
		Updates(ctx, profile)
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}

	profile.UserID = user.ID
	return gorm.G[Profile](database.DB).Create(ctx, &profile)
}
