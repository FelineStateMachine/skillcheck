package tui

import tea "charm.land/bubbletea/v2"

func keyString(msg tea.KeyPressMsg) string { return msg.String() }

func isKey(msg tea.KeyPressMsg, keys ...string) bool {
	got := keyString(msg)
	for _, key := range keys {
		if got == key {
			return true
		}
	}
	return false
}
