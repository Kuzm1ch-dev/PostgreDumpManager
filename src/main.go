package main

import (
	"PostgresDumpManager/src/common"
	"PostgresDumpManager/src/sheduler"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"strings"
)

var DataBaseStorage = common.DataBaseStorage{}
var Config = common.Config{}
var SelectedDB = -1
var SelectedTask = -1

func makeTray(a fyne.App, mainWindow fyne.Window) {
	if desk, ok := a.(desktop.App); ok {
		menu := fyne.NewMenu("MyApp",
			fyne.NewMenuItem("Show", func() {
				mainWindow.Show()
			}))
		desk.SetSystemTrayMenu(menu)
	}
}

func main() {
	sheduler := sheduler.NewSheduler("Europe/Moscow")

	DataBaseStorage.LoadDataBaseFromFile(sheduler)
	Config.LoadConfigFromFile()

	a := app.New()
	w := a.NewWindow("Postgre Log in")
	w.SetCloseIntercept(func() {
		w.Hide()
	})
	makeTray(a, w)

	tabs := container.NewAppTabs(
		container.NewTabItem("PostgreSQL", PostgreDataTab(w)),
		container.NewTabItem("Manager", DataBaseManager(w, sheduler)),
	)

	w.SetContent(
		fyne.NewContainerWithLayout(
			layout.NewMaxLayout(),
			tabs,
		),
	)
	w.ShowAndRun()

}

func PostgreDataTab(w fyne.Window) fyne.CanvasObject {

	HostEntry := widget.NewEntry()
	HostEntry.SetText(os.Getenv("PG_HOST"))
	HostEntry.SetPlaceHolder("Host")

	UserEntry := widget.NewEntry()
	UserEntry.SetPlaceHolder("User")
	UserEntry.SetText(os.Getenv("PG_USER"))

	PasswordEntry := widget.NewPasswordEntry()
	PasswordEntry.SetText(os.Getenv("PG_PASS"))
	PasswordEntry.SetPlaceHolder("Password")

	PostgreBinDir := widget.NewEntry()
	PostgreBinDir.SetText(os.Getenv("PG_DIR"))
	PostgreBinDir.SetPlaceHolder("PostgreSQL bin directory")
	PostgreBinDir.Resize(fyne.Size{100, 32})

	OpenFolder := widget.NewButton("Folder Open", func() {
		dialog.ShowFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if list == nil {
				log.Println("Cancelled")
				return
			}
			PostgreBinDir.SetText(strings.Trim(list.String(), "file://"))
		}, w)
	})

	SaveConfig := widget.NewButton("Save Config", func() {
		Config.Host = HostEntry.Text
		Config.User = UserEntry.Text
		Config.Password = PasswordEntry.Text
		Config.PostgreBinDir = PostgreBinDir.Text
		Config.SaveConfigInFile()
	})

	container := container.NewVBox(
		HostEntry,
		UserEntry,
		PasswordEntry,
		PostgreBinDir,
		OpenFolder,
		SaveConfig,
	)

	return container
}

func DataBaseManager(w fyne.Window, sheduler sheduler.Sheduler) fyne.CanvasObject {
	// List
	DataBaseList := widget.NewList(
		func() int {
			return len(DataBaseStorage.Storage)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.StorageIcon()), widget.NewLabel("Template Object"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(DataBaseStorage.Storage[id].Name)
		},
	)

	DataBaseEntry := widget.NewEntry()
	DataBaseEntry.SetPlaceHolder("DataBase name")
	AddDBButton := widget.NewButton("Add", func() {
		DataBaseStorage.CreateDataBase(DataBaseEntry.Text)
		DataBaseStorage.SaveDataBaseInFile(sheduler)
		DataBaseEntry.SetText("")
		DataBaseList.Refresh()
	})

	EditDBButton := widget.NewButton("Edit", func() {
		if SelectedDB >= 0 {
			TaskManager(SelectedDB, sheduler)
		}
	})

	RemoveDBButton := widget.NewButton("Remove", func() {
		if SelectedDB >= 0 {
			for _, task := range DataBaseStorage.Storage[SelectedDB].Tasks {
				sheduler.RemoveTask(task.EntryID)
			}
			DataBaseStorage.RemoveDataBase(SelectedDB)
		}
		DataBaseList.Refresh()
		DataBaseStorage.SaveDataBaseInFile(sheduler)
	})

	DataBaseManageContainer := container.NewVBox(
		DataBaseEntry,
		AddDBButton,
		EditDBButton,
		RemoveDBButton,
	)

	DataBaseList.OnSelected = func(id widget.ListItemID) {
		SelectedDB = id
		SelectedTask = -1
	}

	container := fyne.NewContainerWithLayout(
		layout.NewGridLayoutWithColumns(2),
		fyne.NewContainerWithLayout(
			layout.NewVBoxLayout(),
			DataBaseManageContainer,
		),
		DataBaseList,
	)

	return container
}

func TaskManager(SelectedDB int, sheduler sheduler.Sheduler) {
	a := fyne.CurrentApp()
	database := DataBaseStorage.GetDataBase(SelectedDB)
	w := a.NewWindow(database.Name)

	TaskList := widget.NewList(
		func() int {
			return len(database.Tasks)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.StorageIcon()), widget.NewLabel("Template Object"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(database.Tasks[id].Name)
		},
	)

	TaskType := widget.NewCard("Task type", "Select the task type",
		widget.NewRadioGroup([]string{"Backup", "Reindex"}, func(s string) {
			database.Tasks[SelectedTask].TaskType = s
		}))

	TaskPeriod := widget.NewCard("Time", "How often to perform the task",
		widget.NewRadioGroup([]string{"Every day", "Every week", "Every month"}, func(s string) {
			database.Tasks[SelectedTask].Period = s
		}))

	TaskTime := widget.NewCard("Repeat every", "",
		widget.NewEntry(),
	)
	TaskTime.Content.(*widget.Entry).PlaceHolder = "h:m:s"

	TaskEntry := widget.NewEntry()
	TaskEntry.SetPlaceHolder("Task name")

	RemoveTaskButton := widget.NewButton("Remove", func() {
		sheduler.RemoveTask(database.Tasks[SelectedTask].EntryID)
		database.RemoveTask(SelectedTask)
		TaskEntry.SetText("")
		DataBaseStorage.SaveDataBaseInFile(sheduler)
		TaskList.Refresh()
	})

	SaveTaskButton := widget.NewButton("Save", func() {

		time := strings.Split(TaskTime.Content.(*widget.Entry).Text, ":")

		err, h := common.CheckTime(time[0], "h")
		if err != nil {
			log.Println(err)
			dialog.ShowInformation("Information", "Incorrect time format", w)
			return
		}
		err, m := common.CheckTime(time[1], "m")
		if err != nil {
			log.Println(err)
			dialog.ShowInformation("Information", "Incorrect time format", w)
			return
		}
		database.Tasks[SelectedTask].Time.H = h
		database.Tasks[SelectedTask].Time.M = m

		sheduler.RemoveTask(database.Tasks[SelectedTask].EntryID)
		entryID, err := sheduler.AddTask(fmt.Sprintf("%d %d * * 0-6", m, h), func() { sheduler.CreateBackUpDataBase(database.Name) })
		if err != nil {
			log.Println(err)
		}
		database.Tasks[SelectedTask].EntryID = entryID
		DataBaseStorage.SaveDataBaseInFile(sheduler)
		TaskList.Refresh()
	})

	TaskEditContainer := container.NewVBox(
		TaskType,
		TaskPeriod,
		TaskTime,
		container.NewHBox(
			SaveTaskButton,
			RemoveTaskButton,
		),
	)

	AddTaskButton := widget.NewButton("Add", func() {
		entryID, err := sheduler.AddTask("0 12 * * 0-6", func() { sheduler.CreateBackUpDataBase("Hello") })
		if err != nil {
			log.Println(err)
		}
		database.CreateTask(TaskEntry.Text, entryID)
		TaskEntry.SetText("")
		TaskList.Refresh()
	})

	TaskCreate := container.NewVBox(
		TaskEntry,
		AddTaskButton,
	)

	container := fyne.NewContainerWithLayout(
		layout.NewHBoxLayout(),
		fyne.NewContainerWithLayout(
			layout.NewGridLayoutWithRows(2),
			TaskCreate,
			TaskList,
		),
		TaskEditContainer,
	)

	TaskList.OnSelected = func(id widget.ListItemID) {
		SelectedTask = id
		TaskType.Content.(*widget.RadioGroup).SetSelected(database.Tasks[SelectedTask].TaskType)
		TaskPeriod.Content.(*widget.RadioGroup).SetSelected(database.Tasks[SelectedTask].Period)
		TaskTime.Content.(*widget.Entry).SetText(database.Tasks[SelectedTask].Time.ToString())
	}

	w.SetContent(
		fyne.NewContainerWithLayout(
			layout.NewMaxLayout(),
			container,
		),
	)
	w.Show()
}
