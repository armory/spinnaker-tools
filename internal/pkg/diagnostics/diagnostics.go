package diagnostics

type Handler interface {
	Action(text string)
	Error(text string, err error)
	UUID() string
	SetEmail(email string)
	IsDiagnostics() bool
}
