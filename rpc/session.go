package rpc

import (
	"sync"
	"time"
)

// Session represents an authenticated device session.
type Session struct {
	// UserID identifies the authenticated user.
	UserID string

	// Data holds arbitrary application-level session data.
	Data map[string]string

	// CreatedAt is the time the session was created.
	CreatedAt time.Time

	// LastActive is the time of the last request from this session.
	LastActive time.Time
}

// SessionStore is the interface for session storage implementations.
// Implementations must be safe for concurrent use.
type SessionStore interface {
	// Get retrieves a session by deviceID. Returns nil if not found.
	Get(deviceID string) (*Session, error)

	// Set stores a session for the given deviceID.
	Set(deviceID string, sess *Session) error

	// Delete removes a session by deviceID.
	Delete(deviceID string) error

	// CleanExpired removes all sessions older than maxAge.
	CleanExpired(maxAge time.Duration) error
}

// MemorySessionStore is an in-memory session store for single-instance deployments.
type MemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*sessionEntry
}

type sessionEntry struct {
	session  *Session
	deviceID string
}

// NewMemorySessionStore creates a new in-memory session store.
func NewMemorySessionStore() *MemorySessionStore {
	return &MemorySessionStore{
		sessions: make(map[string]*sessionEntry),
	}
}

func (m *MemorySessionStore) Get(deviceID string) (*Session, error) {
	m.mu.RLock()
	entry, ok := m.sessions[deviceID]
	m.mu.RUnlock()
	if !ok {
		return nil, nil
	}
	return entry.session, nil
}

func (m *MemorySessionStore) Set(deviceID string, sess *Session) error {
	m.mu.Lock()
	m.sessions[deviceID] = &sessionEntry{session: sess, deviceID: deviceID}
	m.mu.Unlock()
	return nil
}

func (m *MemorySessionStore) Delete(deviceID string) error {
	m.mu.Lock()
	delete(m.sessions, deviceID)
	m.mu.Unlock()
	return nil
}

func (m *MemorySessionStore) CleanExpired(maxAge time.Duration) error {
	threshold := time.Now().Add(-maxAge)
	m.mu.Lock()
	for id, entry := range m.sessions {
		if entry.session.LastActive.Before(threshold) {
			delete(m.sessions, id)
		}
	}
	m.mu.Unlock()
	return nil
}

// StartCleanup starts a background goroutine that periodically cleans expired sessions.
// Returns a stop function that should be called on shutdown.
func (m *MemorySessionStore) StartCleanup(interval, maxAge time.Duration) func() {
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.CleanExpired(maxAge)
			case <-done:
				return
			}
		}
	}()
	return func() { close(done) }
}
