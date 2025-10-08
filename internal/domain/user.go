package domain

import (
	"ivanSaichkin/language-bot/internal/constants"
	"time"
)

type User struct {
	ID           int64               `json:"id"`
	Username     string              `json:"username"`
	FirstName    string              `json:"first_name"`
	LastName     string              `json:"last_name"`
	LanguageCode string              `json:"language_code"`
	State        constants.UserState `json:"state"`
	DailyGoal    int                 `json:"daily_goal"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

func NewUser(userID int64, username, firstName, lastName, languageCode string) *User {
	now := time.Now()
	return &User{
		ID:           userID,
		Username:     username,
		FirstName:    firstName,
		LastName:     lastName,
		LanguageCode: languageCode,
		State:        constants.StateDefault,
		DailyGoal:    10,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (u *User) SetState(state constants.UserState) {
	u.State = state
	u.UpdatedAt = time.Now()
}

func (u *User) SetDailyGoal(goal int) {
	if goal < 1 {
		goal = 1
	}

	if goal > 100 {
		goal = 100
	}

	u.DailyGoal = goal
	u.UpdatedAt = time.Now()
}
