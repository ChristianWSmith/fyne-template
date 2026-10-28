package main

import "fyne.io/fyne/v2"

// Preference keys. fyne.Preferences stores these per app ID in the OS
// settings store (plist on macOS, registry on Windows, dconf on Linux).
const (
	// prefsUpdatesDisabledKey persists the "Never" choice on the update
	// dialog: once true, startup checks stop until the user re-enables
	// them. Help → Check for Updates bypasses the pref (it's the
	// recovery path after choosing Never).
	prefsUpdatesDisabledKey = "updates_disabled"
	// prefsWindowWidthKey / prefsWindowHeightKey persist the window size.
	prefsWindowWidthKey  = "window_width"
	prefsWindowHeightKey = "window_height"
)

// updatesDisabled reports whether the user opted out of update prompts.
func updatesDisabled(prefs fyne.Preferences) bool {
	return prefs.Bool(prefsUpdatesDisabledKey)
}

// savedWindowSize returns the persisted window size when both values are
// usable, else def (a corrupt or out-of-range pref must not break launch).
func savedWindowSize(prefs fyne.Preferences, def fyne.Size) fyne.Size {
	w, h := prefs.Int(prefsWindowWidthKey), prefs.Int(prefsWindowHeightKey)
	if w < 200 || h < 200 || w > 10000 || h > 10000 {
		return def
	}
	return fyne.NewSize(float32(w), float32(h))
}
