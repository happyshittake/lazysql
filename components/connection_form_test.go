package components

import (
	"os"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/jorgerojas26/lazysql/app"
	"github.com/jorgerojas26/lazysql/models"
	"github.com/rivo/tview"
)

func TestConnectionFormClickValidationAndCancel(t *testing.T) {
	for _, label := range []string{"Save", "Test", "Connect", "Cancel"} {
		t.Run(label, func(t *testing.T) {
			form, selection, pages := setupConnectionUI(t, "database = []\n")
			clickConnectionButton(t, selection.Flex, "New")
			clickConnectionButton(t, form.Flex, label)
			if label == "Cancel" {
				if name, _ := pages.GetFrontPage(); name != pageNameConnectionSelection {
					t.Fatal("Cancel did not return to selection")
				}
			} else {
				waitFor(t, func() bool { return form.StatusText.GetText(false) != "" })
				if name, _ := pages.GetFrontPage(); name != pageNameConnectionForm {
					t.Fatal("invalid form should remain open")
				}
			}
			if len(app.App.Connections()) != 0 {
				t.Fatal("invalid/canceled form was saved")
			}
		})
	}
}

func TestConnectionFormEnterOnCancelDoesNotSave(t *testing.T) {
	form, selection, pages := setupConnectionUI(t, "database = []\n")
	clickConnectionButton(t, selection.Flex, "New")
	cancel := clickConnectionButton(t, form.Flex, "Cancel")
	clickConnectionButton(t, selection.Flex, "New")
	app.App.SetFocus(cancel)
	form.Flex.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
	if name, _ := pages.GetFrontPage(); name != pageNameConnectionSelection {
		t.Fatal("Enter on Cancel should cancel, not attempt Save")
	}
	if len(app.App.Connections()) != 0 {
		t.Fatal("Cancel saved a connection")
	}
}

func TestConnectionFormSaveClick(t *testing.T) {
	form, selection, pages := setupConnectionUI(t, "database = []\n")
	clickConnectionButton(t, selection.Flex, "New")
	form.GetFormItem(0).(*tview.InputField).SetText("new")
	form.GetFormItem(1).(*tview.InputField).SetText("sqlite3://new.db")
	clickConnectionButton(t, form.Flex, "Save")
	if name, _ := pages.GetFrontPage(); name != pageNameConnectionSelection {
		t.Fatal("Save did not return to selection")
	}
	connections := app.App.Connections()
	if len(connections) != 1 || connections[0].Name != "new" || connections[0].URL != "sqlite3://new.db" {
		t.Fatalf("saved connections = %+v", connections)
	}
	data, err := os.ReadFile(".lazysql.toml")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "new.db") {
		t.Fatal("Save did not persist connection")
	}
}

func TestConnectionFormDuplicateNames(t *testing.T) {
	for _, edit := range []bool{false, true} {
		t.Run(map[bool]string{false: "create", true: "rename"}[edit], func(t *testing.T) {
			form, selection, pages := setupConnectionUI(t, "[[database]]\nName = 'one'\nURL = 'sqlite3://one.db'\n[[database]]\nName = 'two'\nURL = 'sqlite3://two.db'\n")
			label := "New"
			if edit {
				label = "Edit"
			}
			clickConnectionButton(t, selection.Flex, label)
			form.GetFormItem(0).(*tview.InputField).SetText("two")
			form.GetFormItem(1).(*tview.InputField).SetText("sqlite3://other.db")
			form.Flex.InputHandler()(tcell.NewEventKey(tcell.KeyF1, 0, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
			if name, _ := pages.GetFrontPage(); name != pageNameConnectionForm || !strings.Contains(strings.ToLower(form.StatusText.GetText(false)), "exists") {
				t.Fatal("duplicate name was not rejected")
			}
			connections := app.App.Connections()
			if len(connections) != 2 || connections[0].Name != "one" || connections[1].URL != "sqlite3://two.db" {
				t.Fatal("duplicate name changed existing connections")
			}
		})
	}
}

func TestConnectionFormRenamedEditPreservesFields(t *testing.T) {
	form, selection, _ := setupConnectionUI(t, "[[database]]\nName = 'one'\nURL = 'sqlite3://one.db'\nUsername = 'stored'\nSchemas = ['public']\n[[database.Commands]]\nCommand = 'echo ready'\n")
	clickConnectionButton(t, selection.Flex, "Edit")
	form.GetFormItem(0).(*tview.InputField).SetText("renamed")
	clickConnectionButton(t, form.Flex, "Save")
	connections := app.App.Connections()
	if len(connections) != 1 {
		t.Fatal("editing should update, not append")
	}
	got := connections[0]
	if got.Name != "renamed" || got.Username != "stored" || len(got.Schemas) != 1 || got.Schemas[0] != "public" || len(got.Commands) != 1 || got.Commands[0].Command != "echo ready" {
		t.Fatalf("edit lost fields: %+v", got)
	}
}

func TestConnectionFormConnectSavesBeforeConnecting(t *testing.T) {
	for _, mode := range []string{"click", "F3", "edit"} {
		t.Run(mode, func(t *testing.T) {
			config := "database = []\n"
			label := "New"
			if mode == "edit" {
				config = "[[database]]\nName = 'one'\nURL = 'sqlite3://one.db'\n"
				label = "Edit"
			}
			form, selection, _ := setupConnectionUI(t, config)
			clickConnectionButton(t, selection.Flex, label)
			form.GetFormItem(0).(*tview.InputField).SetText("connected")
			form.GetFormItem(1).(*tview.InputField).SetText("sqlite3://connected.db")
			mainPages.AddPage("connected", tview.NewBox(), true, false)
			startConnectionEventLoop(t)
			app.App.QueueUpdate(func() {
				if mode == "F3" {
					form.Flex.InputHandler()(tcell.NewEventKey(tcell.KeyF3, 0, tcell.ModNone), func(p tview.Primitive) { app.App.SetFocus(p) })
				} else {
					clickConnectionButton(t, form.Flex, "Connect")
				}
			})
			waitFor(t, func() bool {
				connected := false
				app.App.QueueUpdate(func() { name, _ := mainPages.GetFrontPage(); connected = name == "connected" })
				return connected
			})
			connections := app.App.Connections()
			if len(connections) != 1 || connections[0].Name != "connected" || connections[0].URL != "sqlite3://connected.db" {
				t.Fatalf("connection not saved: %+v", connections)
			}
			data, err := os.ReadFile(".lazysql.toml")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(data), "connected.db") {
				t.Fatal("Connect did not persist connection")
			}
		})
	}
}

func TestConnectionFormConnectPersistsBeforeCallback(t *testing.T) {
	form, selection, _ := setupConnectionUI(t, "[[database]]\nName = 'other'\nURL = 'sqlite3://other.db'\n")
	clickConnectionButton(t, selection.Flex, "New")
	form.GetFormItem(0).(*tview.InputField).SetText("new")
	form.GetFormItem(1).(*tview.InputField).SetText("sqlite3://new.db")
	started := false
	form.onConnect = func(connection models.Connection) {
		started = true
		data, err := os.ReadFile(".lazysql.toml")
		if err != nil || !strings.Contains(string(data), "new.db") {
			t.Error("connection started before persistence")
		}
		connections := app.App.Connections()
		row, _ := connectionsTable.GetSelection()
		if len(connections) != 2 || connections[1].Name != "new" || row != 1 || connection.Name != "new" {
			t.Error("Connect did not use the newly saved entry")
		}
	}
	clickConnectionButton(t, form.Flex, "Connect")
	if !started {
		t.Fatal("Connect did not start")
	}
}

func TestConnectionFormSaveOptions(t *testing.T) {
	for _, showAll := range []bool{false, true} {
		t.Run(map[bool]string{false: "pinned", true: "all"}[showAll], func(t *testing.T) {
			form, selection, _ := setupConnectionUI(t, "database = []\n")
			clickConnectionButton(t, selection.Flex, "New")
			form.GetFormItem(0).(*tview.InputField).SetText("new")
			form.GetFormItem(1).(*tview.InputField).SetText("postgres://localhost/mydb")
			form.GetFormItem(2).(*tview.Checkbox).SetChecked(true)
			form.GetFormItem(3).(*tview.Checkbox).SetChecked(showAll)
			clickConnectionButton(t, form.Flex, "Save")
			connections := app.App.Connections()
			if len(connections) != 1 {
				t.Fatal("connection was not saved")
			}
			wantDB := "mydb"
			if showAll {
				wantDB = ""
			}
			if !connections[0].ReadOnly || connections[0].DBName != wantDB {
				t.Fatal("form options were not preserved")
			}
		})
	}
}

func TestConnectionFormConnectSaveFailure(t *testing.T) {
	form, selection, pages := setupConnectionUI(t, "database = []\n")
	clickConnectionButton(t, selection.Flex, "New")
	form.GetFormItem(0).(*tview.InputField).SetText("new")
	form.GetFormItem(1).(*tview.InputField).SetText("sqlite3://new.db")
	if err := os.Remove(".lazysql.toml"); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(".lazysql.toml", 0700); err != nil {
		t.Fatal(err)
	}
	connected := false
	form.onConnect = func(models.Connection) { connected = true }
	clickConnectionButton(t, form.Flex, "Connect")
	if connected {
		t.Fatal("Connect started despite failed save")
	}
	if name, _ := pages.GetFrontPage(); name != pageNameConnectionForm || form.StatusText.GetText(false) == "" {
		t.Fatal("failed save should show an error and stay in form")
	}
	if len(app.App.Connections()) != 0 {
		t.Fatal("failed save changed saved entries")
	}
}

func TestShowAllDatabasesChecked(t *testing.T) {
	testCases := []struct {
		name     string
		conn     models.Connection
		expected bool
	}{
		{
			name: "DBName pinned -> unchecked, regardless of URL",
			conn: models.Connection{
				URL:    "postgres://user:pass@localhost:5432/mydb",
				DBName: "mydb",
			},
			expected: false,
		},
		{
			name: "DBName empty, URL has embedded database -> checked",
			conn: models.Connection{
				URL: "postgres://user:pass@localhost:5432/mydb",
			},
			expected: true,
		},
		{
			name: "DBName empty, URL has no embedded database -> unchecked",
			conn: models.Connection{
				URL: "postgres://user:pass@localhost:5432/",
			},
			expected: false,
		},
		{
			name: "DBName empty, MSSQL URL with embedded database -> checked",
			conn: models.Connection{
				URL: "mssql://user:pass@localhost/DADOSADV",
			},
			expected: true,
		},
		{
			name: "invalid URL -> unchecked (fail safe)",
			conn: models.Connection{
				URL: "not a valid url :: at all",
			},
			expected: false,
		},
		{
			name: "empty connection -> unchecked",
			conn: models.Connection{
				URL: "",
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := showAllDatabasesChecked(tc.conn)
			if got != tc.expected {
				t.Errorf("showAllDatabasesChecked(%+v) = %v, want %v", tc.conn, got, tc.expected)
			}
		})
	}
}
