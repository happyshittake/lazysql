package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/jorgerojas26/lazysql/app"
	"github.com/jorgerojas26/lazysql/drivers"
	"github.com/jorgerojas26/lazysql/helpers"
	"github.com/jorgerojas26/lazysql/internal/telemetry"
	"github.com/jorgerojas26/lazysql/models"
)

type ConnectionForm struct {
	*tview.Flex
	*tview.Form
	StatusText *tview.TextView
	Action     string
	onConnect  func(models.Connection)
}

func NewConnectionForm(connectionPages *models.ConnectionPages) *ConnectionForm {
	wrapper := tview.NewFlex()

	wrapper.SetDirection(tview.FlexColumnCSS)

	addForm := tview.NewForm().SetFieldBackgroundColor(app.Styles.InverseTextColor).SetButtonBackgroundColor(tview.Styles.InverseTextColor).SetLabelColor(tview.Styles.PrimaryTextColor).SetFieldTextColor(tview.Styles.ContrastSecondaryTextColor)
	addForm.AddInputField("Name", "", 0, nil, nil)
	addForm.AddInputField("URL", "", 0, nil, nil)
	addForm.AddCheckbox("Read-Only", false, nil)
	addForm.AddCheckbox("Show all databases in this instance", false, nil)

	buttonsWrapper := tview.NewFlex().SetDirection(tview.FlexColumn)

	saveButton := tview.NewButton(fmt.Sprintf("[%s]F1 [-]Save", app.Styles.SecondaryTextColor))
	saveButton.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimitiveBackgroundColor).Foreground(app.Styles.PrimaryTextColor))
	saveButton.SetBorder(true)

	buttonsWrapper.AddItem(saveButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	testButton := tview.NewButton(fmt.Sprintf("[%s]F2 [-]Test", app.Styles.SecondaryTextColor))
	testButton.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimitiveBackgroundColor).Foreground(app.Styles.PrimaryTextColor))
	testButton.SetBorder(true)

	buttonsWrapper.AddItem(testButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	connectButton := tview.NewButton(fmt.Sprintf("[%s]F3 [-]Connect", app.Styles.SecondaryTextColor))
	connectButton.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimitiveBackgroundColor).Foreground(app.Styles.PrimaryTextColor))
	connectButton.SetBorder(true)

	buttonsWrapper.AddItem(connectButton, 0, 1, false)
	buttonsWrapper.AddItem(nil, 1, 0, false)

	cancelButton := tview.NewButton(fmt.Sprintf("[%s]Esc [-]Cancel", app.Styles.SecondaryTextColor))
	cancelButton.SetStyle(tcell.StyleDefault.Background(app.Styles.PrimitiveBackgroundColor).Foreground(app.Styles.PrimaryTextColor))
	cancelButton.SetBorder(true)

	buttonsWrapper.AddItem(cancelButton, 0, 1, false)

	statusText := tview.NewTextView()
	statusText.SetBorderPadding(1, 1, 0, 0)

	wrapper.AddItem(addForm, 0, 1, true)
	wrapper.AddItem(statusText, 4, 0, false)
	wrapper.AddItem(buttonsWrapper, 3, 0, false)

	form := &ConnectionForm{
		Flex:       wrapper,
		Form:       addForm,
		StatusText: statusText,
	}

	saveButton.SetSelectedFunc(func() { form.runAction(tcell.KeyF1, connectionPages) })
	testButton.SetSelectedFunc(func() { form.runAction(tcell.KeyF2, connectionPages) })
	connectButton.SetSelectedFunc(func() { form.runAction(tcell.KeyF3, connectionPages) })
	cancelButton.SetSelectedFunc(func() { form.runAction(tcell.KeyEsc, connectionPages) })
	wrapper.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter && buttonsWrapper.HasFocus() {
			return event
		}
		switch event.Key() {
		case tcell.KeyEsc, tcell.KeyF1, tcell.KeyEnter, tcell.KeyF2, tcell.KeyF3:
			form.runAction(event.Key(), connectionPages)
			return nil
		}
		return event
	})

	return form
}

func (form *ConnectionForm) runAction(key tcell.Key, connectionPages *models.ConnectionPages) {
	switch key {
	case tcell.KeyEsc:
		connectionPages.SwitchToPage(pageNameConnectionSelection)
	case tcell.KeyF1, tcell.KeyEnter, tcell.KeyF3:
		connection, err := form.saveConnection()
		if err != nil {
			form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(app.Styles.ErrorColor).Background(app.Styles.PrimitiveBackgroundColor))
			return
		}
		connectionPages.SwitchToPage(pageNameConnectionSelection)
		if key == tcell.KeyF3 {
			form.onConnect(connection)
		}
	case tcell.KeyF2:
		connectionString := form.GetFormItem(1).(*tview.InputField).GetText()
		form.testConnection(connectionString)
	}
}

func (form *ConnectionForm) saveConnection() (models.Connection, error) {
	connectionName := form.GetFormItem(0).(*tview.InputField).GetText()
	if connectionName == "" {
		return models.Connection{}, fmt.Errorf("Connection name is required")
	}
	connectionString := form.GetFormItem(1).(*tview.InputField).GetText()
	parsed, err := helpers.ParseConnectionString(connectionString)
	if err != nil {
		return models.Connection{}, err
	}

	databases := app.App.Connections()
	row := len(databases)
	connection := models.Connection{}
	if form.Action == actionEditConnection {
		row, _ = connectionsTable.GetSelection()
		if row < 0 || row >= len(databases) {
			return models.Connection{}, fmt.Errorf("No connection selected")
		}
		// Preserve connection fields that are not exposed by the form.
		connection = databases[row]
	}
	for i, database := range databases {
		if i != row && database.Name == connectionName {
			return models.Connection{}, fmt.Errorf("Connection %q already exists", connectionName)
		}
	}

	var dbName string
	if !form.GetFormItem(3).(*tview.Checkbox).IsChecked() {
		dbName = strings.Split(parsed.Normalize(",", "NULL", 0), ",")[3]
		if dbName == "NULL" {
			dbName = ""
		}
	}
	connection.Name = connectionName
	connection.Provider = parsed.Driver
	connection.DBName = dbName
	connection.URL = connectionString
	connection.ReadOnly = form.GetFormItem(2).(*tview.Checkbox).IsChecked()

	newDatabases := append([]models.Connection(nil), databases...)
	if row == len(databases) {
		newDatabases = append(newDatabases, connection)
	} else {
		newDatabases[row] = connection
	}
	if err := app.App.SaveConnections(newDatabases); err != nil {
		return models.Connection{}, err
	}
	connectionsTable.SetConnections(newDatabases)
	connectionsTable.Select(row, 0)
	return connection, nil
}

func (form *ConnectionForm) testConnection(connectionString string) {
	parsed, err := helpers.ParseConnectionString(connectionString)
	if err != nil {
		usage.ConnectionFailed(telemetry.InvalidConnection)
		form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(app.Styles.ErrorColor).Background(app.Styles.PrimitiveBackgroundColor))
		return
	}

	form.StatusText.SetText("Connecting...").SetTextColor(app.Styles.TertiaryTextColor)

	poolConfig, err := app.App.Config().EffectiveConnectionPool(models.Connection{Provider: parsed.Driver})
	if err != nil {
		form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(tcell.ColorRed))
		return
	}

	var db drivers.Driver

	switch parsed.Driver {
	case drivers.DriverMySQL:
		db = &drivers.MySQL{PoolConfig: poolConfig}
	case drivers.DriverPostgres:
		db = &drivers.Postgres{PoolConfig: poolConfig}
	case drivers.DriverSqlite:
		db = &drivers.SQLite{}
	case drivers.DriverMSSQL:
		db = &drivers.MSSQL{PoolConfig: poolConfig}
	case drivers.DriverClickHouse:
		db = &drivers.ClickHouse{PoolConfig: poolConfig}
	default:
		usage.ConnectionFailed(telemetry.Driver)
		return
	}

	go func() {
		err := db.TestConnection(app.App.Context(), connectionString)
		App.QueueUpdateDraw(func() {
			if err != nil {
				usage.ConnectionFailed(telemetry.ConnectionFailure(err))
				form.StatusText.SetText(err.Error()).SetTextStyle(tcell.StyleDefault.Foreground(app.Styles.ErrorColor).Background(app.Styles.PrimitiveBackgroundColor))
			} else {
				form.StatusText.SetText("Connection success").SetTextColor(app.Styles.TertiaryTextColor)
			}
		})
	}()
}

func (form *ConnectionForm) SetAction(action string) {
	form.Action = action
}

func (form *ConnectionForm) SetConnectionData(conn models.Connection) {
	form.GetFormItem(0).(*tview.InputField).SetText(conn.Name)
	form.GetFormItem(1).(*tview.InputField).SetText(conn.URL)
	form.GetFormItem(2).(*tview.Checkbox).SetChecked(conn.ReadOnly)
	form.GetFormItem(3).(*tview.Checkbox).SetChecked(showAllDatabasesChecked(conn))
}

// showAllDatabasesChecked infers whether the "show all databases" checkbox
// should be pre-checked when editing an existing connection: true when the
// connection has no DBName pinned but its URL does carry an embedded
// database (i.e. it was previously saved with this option enabled, or the
// user cleared DBName by hand). False for freshly created / legacy
// connections where DBName mirrors the URL, preserving prior behavior.
func showAllDatabasesChecked(conn models.Connection) bool {
	if conn.DBName != "" {
		return false
	}

	parsed, err := helpers.ParseConnectionString(conn.URL)
	if err != nil {
		return false
	}

	urlDBName := strings.Split(parsed.Normalize(",", "NULL", 0), ",")[3]

	return urlDBName != "" && urlDBName != "NULL"
}
