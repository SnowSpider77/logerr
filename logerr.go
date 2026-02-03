package logerr

import (
	"fmt"
	"io"
	"time"
)

type LogErr struct {
	class   Class
	message string
	at      time.Time
	err     error
}

func New(class Class, msg string) *LogErr {
	return &LogErr{class: class, message: msg, at: time.Now()}
}

func Wrap(class Class, msg string, err error) *LogErr {
	return &LogErr{class: class, message: msg, at: time.Now(), err: err}
}

func (e *LogErr) Unwrap() error { return e.err }

func (e *LogErr) Class() Class    { return e.class }
func (e *LogErr) Message() string { return e.message }
func (e *LogErr) Time() time.Time { return e.at }

func (e *LogErr) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.err != nil {
		return e.class.String() + ": " + e.message + ": " + e.err.Error()
	}
	return e.class.String() + ": " + e.message
}

func (e *LogErr) LogString() string {
	if e == nil {
		return "<nil>"
	}

	c := GetConfig()
	base := e.class.String() + ": " + e.message
	ts := e.at
	if c.Location != nil {
		ts = ts.In(c.Location)
	}

	switch c.TimePosition {
	case TimeOff:
		return base
	case TimeBefore:
		return ts.Format(c.TimeFormat) + " " + base
	case TimeAfter:
		return base + " " + ts.Format(c.TimeFormat)
	default:
		return base
	}
}

func Log(err error) error {
	if err == nil {
		return nil
	}
	if le, ok := err.(*LogErr); ok {
		le.Log()
		return err
	}

	holder, _ := writer.Load().(*writerHolder)
	w := io.Discard
	if holder != nil && holder.w != nil {
		w = holder.w
	}
	fmt.Fprintln(w, err.Error())
	return err
}

func (e *LogErr) Log() *LogErr {
	if e == nil {
		return nil
	}
	holder, _ := writer.Load().(*writerHolder)
	w := io.Discard
	if holder != nil && holder.w != nil {
		w = holder.w
	}
	fmt.Fprintln(w, e.LogString())
	return e
}
