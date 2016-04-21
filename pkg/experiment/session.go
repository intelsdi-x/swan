package experiment

import "github.com/nu7hatch/gouuid"

type session struct {
	UUID    string
	Name    string
	WorkDir string
}

func sessionNew() session {
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
