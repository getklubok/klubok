package domain

type CookieConfig struct {
	Name        string
	ExpiresDays int
	Domain      string
	Secure      bool
}
