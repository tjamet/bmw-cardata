package bmwcardata

import (
	"context"
	"encoding/json"
	"os"
)

// SessionStore is an interface that allows to store, persist and retrieve authenticated sessions.
// It is used by the Authenticator to store and retrieve the session.
type SessionStore interface {
	Get(ctx context.Context) (*AuthenticatedSession, error)
	Save(ctx context.Context, session *AuthenticatedSession) error
}

// InMemorySessionStore is a session store that stores the session in memory
// It is not persisted and hence a new login worklow will be triggered upon application
// restart.
//
// This is the default session store used by the Authenticator.
type InMemorySessionStore struct {
	session *AuthenticatedSession
}

func (s *InMemorySessionStore) Get(ctx context.Context) (*AuthenticatedSession, error) {
	return s.session, nil
}

func (s *InMemorySessionStore) Save(ctx context.Context, session *AuthenticatedSession) error {
	s.session = session
	return nil
}

// FileSessionStore is a session store that persists the session to a file.
type FileSessionStore struct {
	Path    string
	session *AuthenticatedSession
}

func (s *FileSessionStore) Get(ctx context.Context) (*AuthenticatedSession, error) {
	if s.session != nil {
		return s.session, nil
	}
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, err
	}
	var session AuthenticatedSession
	err = json.Unmarshal(data, &session)
	if err != nil {
		return nil, err
	}
	s.session = &session
	return &session, nil
}

func (s *FileSessionStore) Save(ctx context.Context, session *AuthenticatedSession) error {
	s.session = session
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0600)
}
