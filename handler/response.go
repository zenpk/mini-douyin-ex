package handler

type Response struct {
	StatusCode int32  `json:"status_code"`
	StatusMsg  string `json:"status_msg,omitempty"`
}

const (
	StatusSuccess = 0
	StatusFailed  = 1
)
