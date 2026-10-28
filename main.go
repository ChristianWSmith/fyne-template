package main

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

//go:embed icon.png
var icon []byte

// mainWindow is the single app window; dialogs anchor to it.
var mainWindow fyne.Window

func main() {
	cleanupOlds()

	a := app.NewWithID("com.christianws.fyne-template")
	a.SetIcon(fyne.NewStaticResource("icon.png", icon))
	prefs := a.Preferences()

	w := a.NewWindow("fyne-template")
	mainWindow = w
	w.SetContent(container.NewCenter(container.NewVBox(
		widget.NewLabelWithStyle("Hello, Fyne!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Version "+version, fyne.TextAlignCenter, fyne.TextStyle{}),
	)))
	w.Resize(savedWindowSize(prefs, fyne.NewSize(360, 180)))

	// Persist the window size once on close, not on every drag tick.
	w.SetCloseIntercept(func() {
		size := w.Canvas().Size()
		prefs.SetInt(prefsWindowWidthKey, int(size.Width))
		prefs.SetInt(prefsWindowHeightKey, int(size.Height))
		w.Close()
	})

	w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("Help",
		fyne.NewMenuItem("Check for Updates…", func() { manualUpdateCheck(a) }))))

	go startupUpdateCheck(a)
	w.ShowAndRun()
}

// startupUpdateCheck checks for a newer release on launch. Silent: any
// failure or "no update" is swallowed — a nagging dialog on a network blip
// is worse than missing an update until the next launch. The "Never" pref
// stops the check entirely (Help → Check for Updates bypasses it).
func startupUpdateCheck(a fyne.App) {
	if !shouldCheckUpdate(version, updatesDisabled(a.Preferences()), false, updateAssetName()) {
		return
	}
	checkForUpdate(a, false)
}

// manualUpdateCheck is the Help → Check for Updates action. force bypasses
// the dev-build and "Never" gates so local builds and opted-out users can
// still check, and every outcome (no asset, fetch error, up to date) is
// reported.
func manualUpdateCheck(a fyne.App) {
	checkForUpdate(a, true)
}

// checkForUpdate queries the latest release and, when one is newer than
// the running build, asks the user before installing. Safe from any
// goroutine: all dialogs run on the fyne thread via fyne.Do.
func checkForUpdate(a fyne.App, force bool) {
	if !shouldCheckUpdate(version, updatesDisabled(a.Preferences()), force, updateAssetName()) {
		if force {
			fyne.Do(func() {
				dialog.ShowInformation("No updates", "This platform has no release builds.", mainWindow)
			})
		}
		return
	}
	go func() {
		rel, err := fetchLatestRelease(context.Background(), &http.Client{Timeout: 20 * time.Second})
		fyne.Do(func() {
			if err != nil {
				if force {
					dialog.ShowError(err, mainWindow)
				}
				return
			}
			if !updateAvailable(version, rel.TagName) {
				if force {
					dialog.ShowInformation("Up to date",
						"You already have the latest release ("+version+").", mainWindow)
				}
				return
			}
			showUpdateDialog(a, rel)
		})
	}()
}

// showUpdateDialog offers Update / Later / Never for the pending release.
// Later dismisses and re-offers next launch; Never persists the opt-out
// immediately (a crash must not resurrect the prompt); Update runs the
// download/replace flow behind a progress dialog.
func showUpdateDialog(a fyne.App, rel releaseInfo) {
	content := widget.NewLabel("Version " + rel.TagName + " is available (you have " + version + ").\n\n" +
		"Update downloads the new version, replaces this one, and restarts.")
	content.Wrapping = fyne.TextWrapWord

	var d *dialog.CustomDialog
	laterBtn := widget.NewButton("Later", func() { d.Hide() })
	neverBtn := widget.NewButton("Never", func() {
		// Written immediately, not on app close: the whole point of
		// Never is that it sticks.
		a.Preferences().SetBool(prefsUpdatesDisabledKey, true)
		d.Hide()
	})
	updateBtn := widget.NewButton("Update", func() {
		d.Hide()
		installUpdate(a, rel)
	})
	updateBtn.Importance = widget.SuccessImportance

	buttons := container.NewHBox(laterBtn, neverBtn, layout.NewSpacer(), updateBtn)
	d = dialog.NewCustomWithoutButtons("Update available",
		container.NewBorder(nil, buttons, nil, nil, content), mainWindow)
	d.Resize(fyne.NewSize(420, 0))
	d.Show()
}

// installUpdate downloads and installs rel behind a determinate progress
// dialog: the bar tracks download bytes and the status line walks the
// download → extract → install phases. applyUpdate only returns on
// failure; on success it has already replaced (linux: exec'd in place) or
// spawned (windows/darwin) the new process — this one quits.
func installUpdate(a fyne.App, rel releaseInfo) {
	status := widget.NewLabel("Preparing…")
	bar := widget.NewProgressBar()
	p := dialog.NewCustomWithoutButtons("Updating",
		container.NewVBox(status, bar), mainWindow)
	p.Resize(fyne.NewSize(420, 0))
	p.Show()
	go func() {
		err := applyUpdate(context.Background(), http.DefaultClient, rel, func(u updateProgress) {
			// progressReader already throttles these callbacks (~10/s),
			// so the fyne.Do queue can't be flooded.
			fyne.Do(func() {
				switch u.Phase {
				case "download":
					status.SetText(downloadStatus(u.Written, u.Total))
					if u.Total > 0 {
						bar.SetValue(float64(u.Written) / float64(u.Total))
					}
				case "extract":
					// The download's final tick already filled the bar;
					// extract/install aren't byte-measurable, so only the
					// status line moves from here.
					status.SetText("Extracting…")
				case "install":
					status.SetText("Installing…")
				}
			})
		})
		fyne.Do(func() {
			p.Hide()
			if err != nil {
				dialog.ShowError(fmt.Errorf(
					"%w\n\nIf this keeps failing, download the release manually:\n%s",
					err, updateManualURL), mainWindow)
				return
			}
			a.Quit()
		})
	}()
}

// downloadStatus renders the download phase line, e.g.
// "Downloading… 12.4 MB of 45.6 MB". Total < 0 means unknown length.
func downloadStatus(written, total int64) string {
	if total > 0 {
		return fmt.Sprintf("Downloading… %s of %s", fmtBytes(written), fmtBytes(total))
	}
	return fmt.Sprintf("Downloading… %s", fmtBytes(written))
}

// fmtBytes renders n as a compact human-readable size ("512 B",
// "12.4 MB", "1.2 GB").
func fmtBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGT"[exp])
}
