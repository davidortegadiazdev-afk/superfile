package fileclipboard

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// This integration test replaces the real clipboard. Opt in on a Mac with:
// SUPERFILE_TEST_NATIVE_CLIPBOARD=1 go test ./src/pkg/fileclipboard -run TestMacPasteboard -v
func TestMacPasteboard(t *testing.T) {
	if os.Getenv("SUPERFILE_TEST_NATIVE_CLIPBOARD") != "1" {
		t.Skip("set SUPERFILE_TEST_NATIVE_CLIPBOARD=1 to replace the macOS clipboard")
	}
	dir := t.TempDir()
	// Resolve /var -> /private/var before comparing NSURL paths.
	dir, err := filepath.EvalSymlinks(dir)
	require.NoError(t, err)
	paths := []string{filepath.Join(dir, "a 'quoted' 日本語\nfile.txt"), filepath.Join(dir, "a folder")}
	require.NoError(t, os.WriteFile(paths[0], []byte("clipboard file contents"), 0o600))
	require.NoError(t, os.Mkdir(paths[1], 0o700))
	require.NoError(t, Writer()(paths))
	output, err := exec.Command("/usr/bin/osascript", "-l", "JavaScript", "-e", `
ObjC.import('AppKit');
const urls = $.NSPasteboard.generalPasteboard.readObjectsForClassesOptions($([$.NSURL]), null);
const paths = [];
for (let i = 0; i < urls.count; i++) {
    paths.push(ObjC.unwrap(urls.objectAtIndex(i).path));
}
JSON.stringify(paths);
`).CombinedOutput()
	require.NoError(t, err, string(output))
	var actual []string
	require.NoError(t, json.Unmarshal(output, &actual))
	require.Equal(t, paths, actual, "clipboard must contain native file URLs, not plain path text")
}
