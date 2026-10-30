// Package fileclipboard exports file references to the native system clipboard.
package fileclipboard

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Writer returns the native file clipboard writer, or nil on unsupported systems.
// File references let the receiving application copy the files' contents.
func Writer() func([]string) error {
	if runtime.GOOS != "darwin" {
		return nil
	}
	return writeFiles
}

// Use AppKit's file URL pasteboard representation, as Finder does. Paths are
// passed as arguments, never interpolated into JavaScript or a shell command.
const writeScript = `ObjC.import('AppKit');
function run(paths) {
    const urls = paths.map(function(path) {
        return $.NSURL.fileURLWithPath($(path));
    });
    const pasteboard = $.NSPasteboard.generalPasteboard;
    pasteboard.clearContents;
    if (!pasteboard.writeObjects($(urls))) {
        throw new Error('Could not write files to the macOS clipboard');
    }
}`

const writeTimeout = 5 * time.Second

func writeFiles(paths []string) error {
	return writeWithCommand(paths, func(ctx context.Context, args ...string) ([]byte, error) {
		return exec.CommandContext(ctx, "/usr/bin/osascript", args...).CombinedOutput()
	})
}

func writeWithCommand(paths []string, run func(context.Context, ...string) ([]byte, error)) error {
	if len(paths) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()
	args := []string{"-l", "JavaScript", "-e", writeScript, "--"}
	args = append(args, paths...)
	output, err := run(ctx, args...)
	if err != nil {
		return fmt.Errorf("copy files to macOS clipboard: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}
