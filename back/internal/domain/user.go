package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrUsernameNotAllowed    = errors.New("username not allowed")
	ErrOperationNotAllowed   = errors.New("operation not allowed")
	ErrEmailDomainNotAllowed = errors.New("email domain not allowed")
)

type User struct {
	ID           UserID    `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	ActivateCode *string   `json:"-"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	IsFake       bool      `json:"isFake"`
}

// EmailDomainAllowed возвращает true, если домен email (часть после `@`)
// совпадает с одним из разрешённых суффиксов доменов. Совпадение
// регистронезависимо: равенство суффиксу либо окончание на `.`+суффикс
// (точечная граница). Email без `@` и пустой список суффиксов дают false.
func EmailDomainAllowed(email string, allowedSuffixes []string) bool {
	at := strings.LastIndex(email, "@")
	if at < 0 || at == len(email)-1 {
		return false
	}

	domain := strings.ToLower(email[at+1:])
	for _, suffix := range allowedSuffixes {
		suffix = strings.ToLower(strings.TrimSpace(suffix))
		if suffix == "" {
			continue
		}
		if domain == suffix || strings.HasSuffix(domain, "."+suffix) {
			return true
		}
	}

	return false
}

func IsUsernameDenied(username string) bool {
	username = strings.ToLower(username)

	for _, forbidden := range forbiddenUsernames {
		if forbidden == username || strings.HasPrefix(username, forbidden) || strings.Contains(username, forbidden) {
			return true
		}
	}

	return false
}

var forbiddenUsernames = []string{ //nolint:gochecknoglobals // It's ok
	// Системные роли и доступ:
	"admin",
	"adm",
	"administrator",
	"moder",
	"moderator",
	"support",
	"root",
	"superuser",
	"superadmin",
	"sysadmin",
	"staff",
	"helpdesk",
	"operator",
	"host",
	"server",
	"system",
	"service",
	"bot",
	"robot",
	// Безопасность и защита:
	"security",
	"auth",
	"authentication",
	"login",
	"signin",
	"register",
	"registration",
	"signup",
	"oauth",
	"sso",
	"password",
	"passwd",
	"credential",
	// Технические и системные:
	"api",
	"backend",
	"frontend",
	"dev",
	"developer",
	"development",
	"test",
	"testing",
	"staging",
	"prod",
	"production",
	"localhost",
	"host",
	"server",
	"dns",
	"ftp",
	"ssh",
	"ssl",
	"tls",
	"www",
	"web",
	"website",
	"portal",
	// Управление контентом:
	"owner",
	"master",
	"chief",
	"manager",
	"editor",
	"author",
	"publisher",
	"content",
	"post",
	"blog",
	"news",
	// Юридические и официальные:
	"legal",
	"terms",
	"privacy",
	"policy",
	"tos",
	"agreement",
	"copyright",
	"trademark",
	"abuse",
	"report",
	"complaint",
	// Финансовые:,
	"payment",
	"billing",
	"invoice",
	"purchase",
	"order",
	"cart",
	"checkout",
	"pay",
	"refund",
	"money",
	"fee",
	"price",
	"cost",
	// Пользовательские данные:
	"user",
	"account",
	"profile",
	"me",
	"my",
	"self",
	"anonymous",
	"guest",
	"visitor",
	"member",
	"customer",
	"client",
	// Почтовые и контактные:
	"mail",
	"email",
	"contact",
	"info",
	"support",
	"help",
	"noreply",
	"postmaster",
	"abuse",
	// Резервирование для функциональности:
	"null",
	"nil",
	"undefined",
	"unknown",
	"deleted",
	"removed",
	"banned",
	"blocked",
	"suspended",
	"inactive",
	"void",
	"empty",
}
