package experiment

import "github.com/nu7hatch/gouuid"

type Session struct {
	UUID    string
	Name    string
	WorkDir string
}

func sessionNew() Session {
	session, err := uuid.NewV4()
	if err != nil {
		return Session{}
	}
	return Session{
		UUID:    session.String(),
		Name:    session.String(),
		WorkDir: "",
	}
}
