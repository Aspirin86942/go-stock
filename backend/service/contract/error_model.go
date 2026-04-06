package contract

type Stage string

const (
	StageSource  Stage = "source"
	StageService Stage = "service"
	StageBridge  Stage = "bridge"
)

type UserVisibleError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
	Stage     Stage  `json:"stage"`
}

func NewUserVisibleError(code, message string, retryable bool, stage Stage) UserVisibleError {
	return UserVisibleError{
		Code:      code,
		Message:   message,
		Retryable: retryable,
		Stage:     stage,
	}
}

func (e UserVisibleError) Error() string {
	return e.Message
}
