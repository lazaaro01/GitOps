package model

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *string     `json:"error,omitempty"`
}

type ErrorResponse struct {
	Success bool    `json:"success"`
	Error   string  `json:"error"`
	Data    *string `json:"data,omitempty"`
}
