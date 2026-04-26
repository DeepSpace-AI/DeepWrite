package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID                   string         `json:"id"`
	Email                string         `json:"email"`
	PasswordHash         string         `json:"-"`
	Name                 string         `json:"name"`
	AvatarURL            sql.NullString `json:"avatar_url,omitempty"`
	Institution          sql.NullString `json:"institution,omitempty"`
	ResearchFields       []string       `json:"research_fields,omitempty"`
	SubscriptionTier     string         `json:"subscription_tier"`
	SubscriptionExpiresAt sql.NullTime  `json:"subscription_expires_at,omitempty"`
	EmailVerifiedAt      sql.NullTime   `json:"email_verified_at,omitempty"`
	LastLoginAt          sql.NullTime   `json:"last_login_at,omitempty"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

type Team struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	OwnerID     string    `json:"owner_id"`
	MaxMembers  int       `json:"max_members"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TeamMember struct {
	ID       string    `json:"id"`
	TeamID   string    `json:"team_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type SubscriptionPlan struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Tier          string          `json:"tier"`
	PriceMonthly  sql.NullFloat64 `json:"price_monthly,omitempty"`
	PriceYearly   sql.NullFloat64 `json:"price_yearly,omitempty"`
	Features      []byte          `json:"features"`
	Limits        []byte          `json:"limits"`
	IsActive      bool            `json:"is_active"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type Subscription struct {
	ID                           string       `json:"id"`
	UserID                       string       `json:"user_id"`
	PlanID                       string       `json:"plan_id"`
	Status                       string       `json:"status"`
	CurrentPeriodStart           sql.NullTime `json:"current_period_start,omitempty"`
	CurrentPeriodEnd             sql.NullTime `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd            bool         `json:"cancel_at_period_end"`
	PaymentProvider              sql.NullString `json:"payment_provider,omitempty"`
	PaymentProviderSubscriptionID sql.NullString `json:"payment_provider_subscription_id,omitempty"`
	CreatedAt                    time.Time    `json:"created_at"`
	UpdatedAt                    time.Time    `json:"updated_at"`
}
