package components

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/jorgerojas26/lazysql/app"
	"github.com/jorgerojas26/lazysql/models"
	"github.com/rivo/tview"
)

func setupConnectionUI(t *testing.T, config string) (*ConnectionForm, *ConnectionSelection, *tview.Pages) {
	t.Helper()
	oldApp, oldAlias := app.App, App
	oldPages, oldTable, oldQuit := mainPages, connectionsTable, quitConfirmationModal
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".lazysql.toml"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	app.App = &app.Application{Application: tview.NewApplication()}
	App = app.App
	t.Cleanup(func() {
		app.App, App = oldApp, oldAlias
		mainPages, connectionsTable, quitConfirmationModal = oldPages, oldTable, oldQuit
		if err := os.Chdir(cwd); err != nil {
			t.Error(err)
		}
	})
	if err := app.LoadConfig(filepath.Join(dir, "global.toml")); err != nil {
		t.Fatal(err)
	}
	app.App.Config().ConfirmOnQuit = true
	mainPages = tview.NewPages()
	quitConfirmationModal = nil
	pages := &models.ConnectionPages{Pages: tview.NewPages()}
	form := NewConnectionForm(pages)
	cs := NewConnectionSelection(form, pages)
	pages.AddPage(pageNameConnectionSelection, cs.Flex, true, true)
	pages.AddPage(pageNameConnectionForm, form.Flex, true, false)
	mainPages.AddPage(pageNameConnections, pages.Pages, true, true)
	app.App.SetFocus(mainPages)
	return form, cs, pages.Pages
}

func clickConnectionButton(t *testing.T, root tview.Primitive, label string) *tview.Button {
	t.Helper()
	screen := updateTestScreen(t)
	root.SetRect(0, 0, 100, 24)
	root.Draw(screen)
	var find func(tview.Primitive) *tview.Button
	find = func(p tview.Primitive) *tview.Button {
		if b, ok := p.(*tview.Button); ok && strings.Contains(regexp.MustCompile(`\[[^]]*\]`).ReplaceAllString(b.GetLabel(), ""), label) {
			return b
		}
		if flex, ok := p.(*tview.Flex); ok {
			for i := 0; i < flex.GetItemCount(); i++ {
				if b := find(flex.GetItem(i)); b != nil {
					return b
				}
			}
		}
		return nil
	}
	button := find(root)
	if button == nil {
		t.Fatalf("button %q not found", label)
	}
	x, y, w, h := button.GetRect()
	root.MouseHandler()(tview.MouseLeftDown, tcell.NewEventMouse(x+w/2, y+h/2, tcell.Button1, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
	root.MouseHandler()(tview.MouseLeftClick, tcell.NewEventMouse(x+w/2, y+h/2, tcell.ButtonNone, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
	return button
}

func TestConnectionSelectionClicks(t *testing.T) {
	for _, label := range []string{"New", "Edit", "Delete", "Quit"} {
		t.Run(label, func(t *testing.T) {
			form, selection, pages := setupConnectionUI(t, "[[database]]\nName = 'saved'\nURL = 'sqlite3://test.db'\nProvider = 'sqlite3'\n")
			clickConnectionButton(t, selection.Flex, label)
			if label == "New" || label == "Edit" {
				name, _ := pages.GetFrontPage()
				if name != pageNameConnectionForm {
					t.Fatalf("%s click did not open form", label)
				}
				want := ""
				if label == "Edit" {
					want = "saved"
				}
				if got := form.GetFormItem(0).(*tview.InputField).GetText(); got != want {
					t.Fatalf("name = %q, want %q", got, want)
				}
			} else if !mainPages.HasPage(pageNameConfirmation) {
				t.Fatalf("%s click did not request confirmation", label)
			}
			if len(app.App.Connections()) != 1 {
				t.Fatal("click changed saved entries without confirmation")
			}
		})
	}
}

func startConnectionEventLoop(t *testing.T) {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	screen.SetSize(100, 24)
	app.App.Application.SetScreen(screen).SetRoot(mainPages, true)
	application := app.App.Application
	done := make(chan error, 1)
	go func() { done <- application.Run() }()
	application.QueueUpdate(func() {})
	t.Cleanup(func() {
		application.Stop()
		if err := <-done; err != nil {
			t.Error(err)
		}
	})
}

func TestConnectionSelectionConnectClick(t *testing.T) {
	_, selection, _ := setupConnectionUI(t, "[[database]]\nName = 'saved'\nURL = 'sqlite3://test.db'\nProvider = 'sqlite3'\n")
	mainPages.AddPage("saved", tview.NewBox(), true, false)
	startConnectionEventLoop(t)
	app.App.QueueUpdate(func() { clickConnectionButton(t, selection.Flex, "Connect") })
	waitFor(t, func() bool {
		connected := false
		app.App.QueueUpdate(func() { name, _ := mainPages.GetFrontPage(); connected = name == "saved" })
		return connected
	})
}

func TestConnectionSelectionFailedDeleteKeepsRows(t *testing.T) {
	_, selection, _ := setupConnectionUI(t, "[[database]]\nName = 'one'\n[[database]]\nName = 'two'\n[[database]]\nName = 'three'\n")
	clickConnectionButton(t, selection.Flex, "Delete")
	if err := os.Remove(".lazysql.toml"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(".lazysql.toml", 0700); err != nil {
		t.Fatal(err)
	}
	_, modal := mainPages.GetFrontPage()
	modal.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
	connections := connectionsTable.GetConnections()
	if len(connections) != 3 || connections[0].Name != "one" || connections[1].Name != "two" || connections[2].Name != "three" {
		t.Fatalf("failed delete changed row targets: %+v", connections)
	}
}

func TestConnectionFormConnectDoesNotReuseChangedOpenSession(t *testing.T) {
	form, selection, pages := setupConnectionUI(t, "[[database]]\nName = 'one'\nURL = 'sqlite3://one.db'\n")
	selection.connected["one"] = app.App.Connections()[0]
	home := &Home{Flex: tview.NewFlex(), ConnectionURL: "sqlite3://one.db", ReadOnly: false}
	mainPages.AddPage("one", home, true, false)
	clickConnectionButton(t, selection.Flex, "Edit")
	form.GetFormItem(2).(*tview.Checkbox).SetChecked(true)
	startConnectionEventLoop(t)
	app.App.QueueUpdate(func() { clickConnectionButton(t, form.Flex, "Connect") })
	// Drain pending UI work before checking the page and status.
	waitFor(t, func() bool {
		settled := false
		app.App.QueueUpdate(func() {
			name, _ := mainPages.GetFrontPage()
			settled = name == "one" || selection.StatusText.GetText(false) != ""
		})
		return settled
	})
	app.App.QueueUpdate(func() {
		if name, _ := pages.GetFrontPage(); name != pageNameConnectionSelection {
			t.Error("edited connection did not return to selection")
		}
		if name, _ := mainPages.GetFrontPage(); name == "one" || selection.StatusText.GetText(false) == "" {
			t.Error("Connect silently reused a writable session after saving Read-Only")
		}
	})
	if !app.App.Connections()[0].ReadOnly {
		t.Fatal("edited settings were not saved")
	}
}

func TestConnectionSelectionEmptyListClicks(t *testing.T) {
	_, selection, pages := setupConnectionUI(t, "database = []\n")
	for _, label := range []string{"Edit", "Delete", "Connect"} {
		clickConnectionButton(t, selection.Flex, label)
	}
	if name, _ := pages.GetFrontPage(); name != pageNameConnectionSelection || mainPages.HasPage(pageNameConfirmation) {
		t.Fatal("selected-entry actions should do nothing without a connection")
	}
}
