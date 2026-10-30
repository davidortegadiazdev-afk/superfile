package internal

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yorukot/superfile/src/internal/common"
	"github.com/yorukot/superfile/src/pkg/utils"
)

func TestCopyItemsExportsNativeFiles(t *testing.T) {
	dir := t.TempDir()
	first, second := filepath.Join(dir, "a file.txt"), filepath.Join(dir, "folder")
	utils.SetupFiles(t, first)
	utils.SetupDirectories(t, second)

	for _, multiple := range []bool{false, true} {
		t.Run(map[bool]string{false: "focused file", true: "multiple files and folders"}[multiple], func(t *testing.T) {
			m := defaultTestModel(dir)
			panel := m.getFocusedFilePanel()
			setFilePanelSelectedItemByLocation(t, panel, first)
			if multiple {
				panel.ChangeFilePanelMode()
				panel.SetSelected(first)
				panel.SetSelected(second)
			}
			var copied []string
			m.fileClipboardWriter = func(paths []string) error { copied = paths; return nil }
			cmd := m.normalAndBrowserModeKey(common.Hotkeys.CopyItems[0])
			require.NotNil(t, cmd)
			require.Nil(t, copied, "native clipboard must be written asynchronously")
			expected := append([]string(nil), m.clipboard.GetItems()...)
			if multiple {
				require.ElementsMatch(t, []string{first, second}, expected)
			} else {
				require.Equal(t, []string{first}, expected)
			}
			m.clipboard.Reset(false)
			require.Nil(t, cmd())
			require.Equal(t, expected, copied, "command must capture the original selection")
		})
	}
}

func TestNativeClipboardFailureKeepsInternalCopy(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	utils.SetupFiles(t, file)
	m := defaultTestModel(dir)
	m.fileClipboardWriter = func([]string) error { return errors.New("not available") }
	cmd := m.normalAndBrowserModeKey(common.Hotkeys.CopyItems[0])
	require.NotNil(t, cmd)
	msg, ok := cmd().(NotifyModalUpdateMsg)
	require.True(t, ok)
	msg.ApplyToModel(m)
	require.True(t, m.notifyModel.IsOpen())
	require.Contains(t, m.notifyModel.GetContent(), "not available")
	require.Equal(t, []string{file}, m.clipboard.GetItems())
}

func TestCutDoesNotExportNativeCopy(t *testing.T) {
	dir := t.TempDir()
	utils.SetupFiles(t, filepath.Join(dir, "file.txt"))
	m := defaultTestModel(dir)
	m.fileClipboardWriter = func([]string) error { t.Fatal("cut must not export a copy"); return nil }
	require.Nil(t, m.normalAndBrowserModeKey(common.Hotkeys.CutItems[0]))
	require.True(t, m.clipboard.IsCut())
}
