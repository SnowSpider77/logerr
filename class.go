package logerr

type Class uint8

const (
	Unknown Class = iota
	Info
	Warning
	Error
	Critical
)

func (c Class) String() string {
	switch c {
	case Info:
		return "info"
	case Warning:
		return "warning"
	case Error:
		return "error"
	case Critical:
		return "critical"
	default:
		return "unknown"
	}
}
