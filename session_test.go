package bmwcardata

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemorySessionManager_SaveGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	m := &InMemorySessionStore{}

	// Initially empty
	s, err := m.Get(ctx)
	require.NoError(t, err)
	assert.Nil(t, s)

	// Save and then Get
	expiresAt := time.Now().Add(1 * time.Hour)
	want := &AuthenticatedSession{AccessToken: "access", RefreshToken: "refresh", ExpiresAt: expiresAt}
	require.NoError(t, m.Save(ctx, want))
	got, err := m.Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, want.AccessToken, got.AccessToken)
	assert.Equal(t, want.RefreshToken, got.RefreshToken)
	assert.Equal(t, want.ExpiresAt, got.ExpiresAt)
}

func TestFileSessionManager_SaveGet_ReadsFromDiskAndCaches(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dir := t.TempDir()
	path := filepath.Join(dir, "session.json")

	// Save a session
	writer := &FileSessionStore{Path: path}
	expiresAt := time.Now().Add(2 * time.Hour)
	stored := &AuthenticatedSession{AccessToken: "tok1", RefreshToken: "ref1", ExpiresAt: expiresAt}
	modified := &AuthenticatedSession{AccessToken: "tok2", RefreshToken: "ref2", ExpiresAt: expiresAt}
	require.NoError(t, writer.Save(ctx, stored))

	// New manager reads from disk (not cache of writer instance)
	reader := &FileSessionStore{Path: path}
	got, err := reader.Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, stored.AccessToken, got.AccessToken)
	assert.Equal(t, stored.RefreshToken, got.RefreshToken)

	// After first Get() the reader caches; modifying the file should not affect returned value
	require.NoError(t, writer.Save(ctx, modified))
	cached, err := reader.Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, stored.AccessToken, cached.AccessToken)
	assert.Equal(t, stored.RefreshToken, cached.RefreshToken)
}
