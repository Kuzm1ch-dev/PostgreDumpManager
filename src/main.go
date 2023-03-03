package main

import (
	"PostgresDumpManager/src/common"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"log"
	"strings"
)

var DataBases = []common.DataBase{}
var SelectedDB = -1
var SelectedTask = -1

func main() {
	DataBases = common.Load()

	a := app.New()
	w := a.NewWindow("Postgre Log in")
	w.FullScreen()

	tabs := container.NewAppTabs(
		container.NewTabItem("PostgreSQL", PostgreDataTab(w)),
		container.NewTabItem("Manager", DataBaseManager(w)),
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

func DataBaseManager(w fyne.Window) fyne.CanvasObject {
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
	AddDBButton := widget.NewButton("Добавить", func() {
		DataBases = append(DataBases, common.DataBase{DataBaseEntry.Text, []common.Task{}})
		DataBaseEntry.SetText("")
		DataBaseList.Refresh()
		common.Save(DataBases)
	})

	EditDBButton := widget.NewButton("Редактировать", func() {
		if SelectedDB >= 0 {
			TaskManager(SelectedDB)
		}
	})

	DataBaseManageContainer := container.NewVBox(
		DataBaseEntry,
		AddDBButton,
		EditDBButton,
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

func TaskManager(SelectedDB int) {
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
	TaskEntry.SetPlaceHolder("Имя задачи")

	RemoveTaskButton := widget.NewButton("Remove", func() {
		DataBases[SelectedDB].Tasks = append(DataBases[SelectedDB].Tasks[:SelectedTask], DataBases[SelectedDB].Tasks[SelectedTask+1:]...)
		TaskEntry.SetText("")
		common.Save(DataBases)
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
		err, s := common.CheckTime(time[2], "s")
		if err != nil {
			log.Println(err)
			dialog.ShowInformation("Information", "Incorrect time format", w)
			return
		}
		DataBases[SelectedDB].Tasks[SelectedTask].Time.H = h
		DataBases[SelectedDB].Tasks[SelectedTask].Time.M = m
		DataBases[SelectedDB].Tasks[SelectedTask].Time.S = s

		common.Save(DataBases)
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
		DataBases[SelectedDB].Tasks = append(DataBases[SelectedDB].Tasks, common.Task{TaskEntry.Text, "Backup", "Every day", common.Time{12, 0, 0}})
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
