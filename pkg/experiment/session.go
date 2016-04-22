package experiment

import "github.com/nu7hatch/gouuid"

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
		Name:    s.String(),
		WorkDir: "",
	}
}
