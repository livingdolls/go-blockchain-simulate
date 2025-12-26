package dto

type APIResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewSuccessResponse[T any](data T) APIResponse[T] {
	return APIResponse[T]{
		Success: true,
		Data:    data,
	}
}

func NewErrorResponse[T any](errMsg string) APIResponse[T] {
	return APIResponse[T]{
		Success: false,
		Error:   errMsg,
	}
}
