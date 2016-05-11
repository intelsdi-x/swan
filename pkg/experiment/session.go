package experiment

import (
	"github.com/nu7hatch/gouuid"
	"time"
)

type session struct {
	UUID    string
	Name    string
	WorkDir string
}

func newSession() session {
	s, err := uuid.NewV4()
	if err != nil {
		return session{}
	}
	return session{
		UUID:    s.String(),
		Name:    time.Now().Format("2006-01-02T15h04m05s_") + s.String(),
		WorkDir: "",
	}
}
