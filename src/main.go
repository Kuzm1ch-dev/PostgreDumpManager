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
	"strings"
)

var DataBases = []common.DataBase{}
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
	DataBases = common.Load(sheduler)
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
	HostEntry.SetPlaceHolder("Host")

	UserEntry := widget.NewEntry()
	UserEntry.SetPlaceHolder("User")

	PasswordEntry := widget.NewEntry()
	PasswordEntry.SetPlaceHolder("Password")

	PostgreBinDir := widget.NewEntry()
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

	container := container.NewVBox(
		HostEntry,
		UserEntry,
		PasswordEntry,
		PostgreBinDir,
		OpenFolder,
	)

	return container
}

func DataBaseManager(w fyne.Window, sheduler sheduler.Sheduler) fyne.CanvasObject {
	// List
	DataBaseList := widget.NewList(
		func() int {
			return len(DataBases)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.StorageIcon()), widget.NewLabel("Template Object"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(DataBases[id].Name)
		},
	)

	DataBaseEntry := widget.NewEntry()
	DataBaseEntry.SetPlaceHolder("DataBase name")
	AddDBButton := widget.NewButton("Add", func() {
		DataBases = append(DataBases, common.DataBase{DataBaseEntry.Text, []common.Task{}})
		DataBaseEntry.SetText("")
		DataBaseList.Refresh()
		common.Save(DataBases, sheduler)
	})

	EditDBButton := widget.NewButton("Edit", func() {
		if SelectedDB >= 0 {
			TaskManager(SelectedDB, sheduler)
		}
	})

	RemoveDBButton := widget.NewButton("Remove", func() {
		if SelectedDB >= 0 {
			for _, task := range DataBases[SelectedDB].Tasks {
				sheduler.RemoveTask(task.EntryID)
			}
			DataBases = append(DataBases[:SelectedDB], DataBases[SelectedDB+1:]...)
		}
		DataBaseList.Refresh()
		common.Save(DataBases, sheduler)
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
	w := a.NewWindow(DataBases[SelectedDB].Name)

	TaskList := widget.NewList(
		func() int {
			return len(DataBases[SelectedDB].Tasks)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.StorageIcon()), widget.NewLabel("Template Object"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(DataBases[SelectedDB].Tasks[id].Name)
		},
	)

	TaskType := widget.NewCard("Task type", "Select the task type",
		widget.NewRadioGroup([]string{"Backup", "Reindex"}, func(s string) {
			DataBases[SelectedDB].Tasks[SelectedTask].TaskType = s
		}))

	TaskPeriod := widget.NewCard("Time", "How often to perform the task",
		widget.NewRadioGroup([]string{"Every day", "Every week", "Every month"}, func(s string) {
			DataBases[SelectedDB].Tasks[SelectedTask].Period = s
		}))

	TaskTime := widget.NewCard("Repeat every", "",
		widget.NewEntry(),
	)
	TaskTime.Content.(*widget.Entry).PlaceHolder = "h:m:s"

	TaskEntry := widget.NewEntry()
	TaskEntry.SetPlaceHolder("Task name")

	RemoveTaskButton := widget.NewButton("Remove", func() {
		sheduler.RemoveTask(DataBases[SelectedDB].Tasks[SelectedTask].EntryID)
		DataBases[SelectedDB].Tasks = append(DataBases[SelectedDB].Tasks[:SelectedTask], DataBases[SelectedDB].Tasks[SelectedTask+1:]...)
		TaskEntry.SetText("")
		common.Save(DataBases, sheduler)
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
		DataBases[SelectedDB].Tasks[SelectedTask].Time.H = h
		DataBases[SelectedDB].Tasks[SelectedTask].Time.M = m

		sheduler.RemoveTask(DataBases[SelectedDB].Tasks[SelectedTask].EntryID)
		entryID, err := sheduler.AddTask(fmt.Sprintf("%d %d * * 0-6", m, h), func() { sheduler.CreateBackUpDataBase("Hello") })
		if err != nil {
			log.Println(err)
		}
		DataBases[SelectedDB].Tasks[SelectedTask].EntryID = entryID
		common.Save(DataBases, sheduler)
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
		DataBases[SelectedDB].Tasks = append(DataBases[SelectedDB].Tasks, common.Task{TaskEntry.Text, "Backup", "Every day", common.Time{12, 0}, entryID})
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
		TaskType.Content.(*widget.RadioGroup).SetSelected(DataBases[SelectedDB].Tasks[SelectedTask].TaskType)
		TaskPeriod.Content.(*widget.RadioGroup).SetSelected(DataBases[SelectedDB].Tasks[SelectedTask].Period)
		TaskTime.Content.(*widget.Entry).SetText(DataBases[SelectedDB].Tasks[SelectedTask].Time.ToString())
	}

	w.SetContent(
		fyne.NewContainerWithLayout(
			layout.NewMaxLayout(),
			container,
		),
	)
	w.Show()
}
