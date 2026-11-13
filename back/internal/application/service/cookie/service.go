package cookie

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"klubok/internal/domain"
)

type Service struct {
	cookieConfig domain.CookieConfig

	nowFunc func() time.Time
}

func NewService(cookieConfig domain.CookieConfig) *Service {
	return &Service{
		cookieConfig: cookieConfig,

		nowFunc: time.Now,
	}
}

func (s *Service) Create(v domain.CookieID) http.Cookie {
	return http.Cookie{
		Name:     s.cookieConfig.Name,
		Value:    v.String(),
		Path:     "/",
		Domain:   s.cookieConfig.Domain,
		Expires:  s.nowFunc().UTC().AddDate(0, 0, s.cookieConfig.ExpiresDays),
		Secure:   s.cookieConfig.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}

func (s *Service) Get(r *http.Request) (domain.CookieID, error) {
	c, err := r.Cookie(s.cookieConfig.Name)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return "", http.ErrNoCookie
		}

		return "", fmt.Errorf("get cookie: %w", err)
	}

	return domain.CookieID(c.Value), nil
}

func (s *Service) Clear() http.Cookie {
	return http.Cookie{
		Name:     s.cookieConfig.Name,
		Value:    "",
		Path:     "/",
		Domain:   s.cookieConfig.Domain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   s.cookieConfig.Secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
}
