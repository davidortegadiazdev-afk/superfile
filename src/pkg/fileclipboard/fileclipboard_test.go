package fileclipboard

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteFilesPassesPathsAsArguments(t *testing.T) {
	paths := []string{"/tmp/a file.txt", "/tmp/quote'\";$(touch injected)\n日本語", "/tmp/folder"}
	err := writeWithCommand(paths, func(ctx context.Context, args ...string) ([]byte, error) {
		_, hasDeadline := ctx.Deadline()
		require.True(t, hasDeadline)
		require.Equal(t, []string{"-l", "JavaScript", "-e", writeScript, "--"}, args[:5])
		require.Equal(t, paths, args[5:])
		return nil, nil
	})
	require.NoError(t, err)
}

func TestWriteFilesEmptyDoesNotClearClipboard(t *testing.T) {
	require.NoError(t, writeWithCommand(nil, func(context.Context, ...string) ([]byte, error) {
		t.Fatal("empty selection must not invoke osascript")
		return nil, nil
	}))
}

func TestWriteFilesReportsCommandFailure(t *testing.T) {
	failure := errors.New("exit status 1")
	err := writeWithCommand([]string{"/tmp/file"}, func(context.Context, ...string) ([]byte, error) {
		return []byte("pasteboard unavailable\n"), failure
	})
	require.ErrorIs(t, err, failure)
	require.ErrorContains(t, err, "pasteboard unavailable")
}
