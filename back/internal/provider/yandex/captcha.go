package yandex

type CaptchaResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Host    string `json:"host"`
}
