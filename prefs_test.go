package main

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestSavedWindowSize(t *testing.T) {
	prefs := test.NewTempApp(t).Preferences()
	def := fyne.NewSize(360, 180)
	if got := savedWindowSize(prefs, def); got != def {
		t.Errorf("empty prefs = %v, want default %v", got, def)
	}
	prefs.SetInt(prefsWindowWidthKey, 800)
	prefs.SetInt(prefsWindowHeightKey, 600)
	if got := savedWindowSize(prefs, def); got != fyne.NewSize(800, 600) {
		t.Errorf("saved size = %v, want 800x600", got)
	}
	// Corrupt / out-of-range values fall back to the default rather than
	// opening a zero-sized or absurdly large window.
	prefs.SetInt(prefsWindowWidthKey, 0)
	if got := savedWindowSize(prefs, def); got != def {
		t.Errorf("zero width = %v, want default %v", got, def)
	}
	prefs.SetInt(prefsWindowWidthKey, 800)
	prefs.SetInt(prefsWindowHeightKey, 999999)
	if got := savedWindowSize(prefs, def); got != def {
		t.Errorf("absurd height = %v, want default %v", got, def)
	}
}

func TestUpdatesDisabled(t *testing.T) {
	prefs := test.NewTempApp(t).Preferences()
	if updatesDisabled(prefs) {
		t.Error("default pref should be false")
	}
	prefs.SetBool(prefsUpdatesDisabledKey, true)
	if !updatesDisabled(prefs) {
		t.Error("pref set to true should read back true")
	}
}
