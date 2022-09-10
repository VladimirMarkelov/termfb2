package main

import (
	"flag"
	"fmt"
	ui "github.com/VladimirMarkelov/clui"
	fbutils "github.com/VladimirMarkelov/fb2text"
	"github.com/VladimirMarkelov/termfb2/common"
	cf "github.com/VladimirMarkelov/termfb2/config"
	xs "github.com/huandu/xstrings"
	term "github.com/nsf/termbox-go"
	"os"
	path "path/filepath"
	"unicode/utf8"
)

const (
	CONFIGFILE   = "config"
	LASTFILE     = "last"
	DBFILE       = "book.db"
	DBCOLLECTION = "books"
	VENDOR       = ".rionnag"
	APPNAME      = "termfb2"

    minWidth = 25
    minHeight = 14
)

// ControlList is a list of all UI widgets that are managed by
// the application
type ControlList struct {
	// The very main Window widget
	mainWindow *ui.Window
	// The widget to display a book text
	reader *ui.TextReader
	// The book library - used only if book library is ON (default is ON)
	bookListWindow *ui.Window
	// The TableView widget that shows book library
	bookTable *ui.TableView
	// The bottom widget to display full text of the currently selected
	// table cell
	bookInfoDetail *ui.Label

	// Widgets for delete book confirmation dialog
	// used only in case of book library is ON
	askWindow *ui.Window
	askLabel  *ui.Label
	askRemove *ui.Button
	askCancel *ui.Button
}

// createView creates the main Window - a book reader view
func createView(controls *ControlList, conf *cf.Config) {
	controls.mainWindow = ui.AddWindow(0, 0, 12, 7, "TermFB2")
	controls.mainWindow.SetPack(ui.Vertical)

	controls.mainWindow.OnKeyDown(func(ev ui.Event, data interface {}) bool {
		if ev.Key == term.KeyF2 && conf.UseDb {
			createBookListDialog(controls, conf)
			return true
		}
		return false
	}, nil)
	controls.reader = ui.CreateTextReader(controls.mainWindow, minWidth, minHeight, 1)
	controls.reader.SetTextColor(conf.TextColor)
	controls.reader.SetBackColor(conf.BackColor)
	ui.ActivateControl(controls.mainWindow, controls.reader)
	controls.mainWindow.SetMaximized(true)
	controls.mainWindow.SetModal(true)
}

func mainLoop(controls *ControlList, conf *cf.Config) {
	ui.MainLoop()
}

// Generate a text for TableView control that displays book library
func getBookColumnText(book common.BookRecord, col int) string {
	text := ""
	switch col {
	case 0:
		text = book.LastName + ", " + book.FirstName
	case 1:
		text = book.Title
	case 2:
		if book.LineTotal == 0 {
			text = "0%"
		} else {
			text = fmt.Sprintf("%v%%", book.LineLast*100/book.LineTotal)
		}
	case 3:
		text = book.Sequence
	case 4:
		text = book.Genre
	case 5:
		text = book.Added
	case 6:
		text = book.Completed
	case 7:
		text = book.FilePath
	}

	return text
}

// Opens a book that is currently selected in TableView (book library)
func loadBook(controls *ControlList, conf *cf.Config) {
	var lines []string
	row := controls.bookTable.SelectedRow()
	b := conf.DbDriver.FilteredBooks()[row]
	fileName := b.FilePath

	conf.Info, lines = fbutils.ParseBook(fileName, true)
	width, _ := controls.reader.Size()
	conf.Lines = fbutils.FormatBook(lines, width, conf.Justify)
	lastPosition := b.LineLast
	conf.LastLength = len(conf.Lines)
	conf.LastFile = fileName

	if lastPosition >= conf.LastLength {
		lastPosition = conf.LastLength - 1
	}
	controls.reader.SetLineCount(conf.LastLength)
	conf.LastPosition = lastPosition
	controls.reader.SetTopLine(conf.LastPosition)
}

// Creates a confirmation dialog to use it when asking about
// deleting a book from the library
func createBookConfirm(controls *ControlList, conf *cf.Config) {
	cw, ch := term.Size()
	halfWidth := cw / 2
	dlgWidth := (halfWidth - 5) * 2
	controls.askWindow = ui.AddWindow(5, ch/2-8, dlgWidth, 3, "Remove book")
	controls.askWindow.SetConstraints(dlgWidth, ui.KeepValue)
	controls.askWindow.SetModal(false)
	controls.askWindow.SetPack(ui.Vertical)

	ui.CreateFrame(controls.askWindow, 1, 1, ui.BorderNone, ui.Fixed)
	fbtn := ui.CreateFrame(controls.askWindow, 1, 1, ui.BorderNone, 1)
	ui.CreateFrame(fbtn, 1, 1, ui.BorderNone, ui.Fixed)
	controls.askLabel = ui.CreateLabel(fbtn, 10, 3, "Remove book?", 1)
	controls.askLabel.SetMultiline(true)
	ui.CreateFrame(fbtn, 1, 1, ui.BorderNone, ui.Fixed)

	ui.CreateFrame(controls.askWindow, 1, 1, ui.BorderNone, ui.Fixed)
	frm1 := ui.CreateFrame(controls.askWindow, 16, 4, ui.BorderNone, ui.Fixed)
	ui.CreateFrame(frm1, 1, 1, ui.BorderNone, 1)
	controls.askRemove = ui.CreateButton(frm1, ui.AutoSize, ui.AutoSize, "Remove", ui.Fixed)
	controls.askCancel = ui.CreateButton(frm1, ui.AutoSize, ui.AutoSize, "Cancel", ui.Fixed)
	controls.askCancel.OnClick(func(ev ui.Event) {
		controls.askWindow.SetModal(false)
		controls.askWindow.SetVisible(false)
		ui.ActivateControl(controls.bookListWindow, controls.bookTable)
	})

	controls.askWindow.SetVisible(false)
	controls.askWindow.OnClose(func(ev ui.Event) bool {
		controls.askWindow.SetVisible(false)
		controls.bookListWindow.SetModal(true)
		ui.ActivateControl(controls.bookListWindow, controls.bookTable)
		return false
	})
}

// Creates and shows a book library dialog - available only if
// library is ON
func createBookListDialog(controls *ControlList, conf *cf.Config) {
	controls.bookListWindow = ui.AddWindow(0, 0, 12, 7, "Book list")
	controls.bookListWindow.SetPack(ui.Vertical)
	controls.bookListWindow.SetModal(true)

	controls.bookTable = ui.CreateTableView(controls.bookListWindow, minWidth, minHeight, 1)
	controls.bookInfoDetail = ui.CreateLabel(controls.bookListWindow, 1, 1, "", ui.Fixed)
	ui.ActivateControl(controls.bookListWindow, controls.bookTable)
	controls.bookTable.SetShowLines(true)
	controls.bookTable.SetShowRowNumber(true)
	controls.bookListWindow.SetMaximized(true)

	controls.bookTable.SetRowCount(len(conf.DbDriver.FilteredBooks()))
	controls.bookListWindow.SetTitle(fmt.Sprintf("Book list [%s]", conf.DbDriver.Filter()))

	cols := []ui.Column{
		ui.Column{Title: "Author", Width: 16, Alignment: ui.AlignLeft},
		ui.Column{Title: "Title", Width: 25, Alignment: ui.AlignLeft},
		ui.Column{Title: "Done", Width: 4, Alignment: ui.AlignRight},
		ui.Column{Title: "Sequence", Width: 8, Alignment: ui.AlignLeft},
		ui.Column{Title: "Genre", Width: 8, Alignment: ui.AlignLeft},
		ui.Column{Title: "Added", Width: 20, Alignment: ui.AlignLeft},
		ui.Column{Title: "Completed", Width: 20, Alignment: ui.AlignLeft},
		ui.Column{Title: "FilePath", Width: 100, Alignment: ui.AlignLeft},
	}
	controls.bookTable.SetColumns(cols)

	// override OnKeyDown to support incremental search and
	// opening selected book by pressing Enter
	// Escape closes the dialog without doing anything
	controls.bookListWindow.OnKeyDown(func(ev ui.Event, data interface {}) bool {
		if ev.Ch != 0 {
			filter := conf.DbDriver.Filter() + string(ev.Ch)
			conf.DbDriver.SetFilter(filter)

			controls.bookTable.SetRowCount(len(conf.DbDriver.FilteredBooks()))
			controls.bookListWindow.SetTitle(fmt.Sprintf("Book list [%s]", filter))
			return true
		}

		switch ev.Key {
		case term.KeyBackspace:
			filter := conf.DbDriver.Filter()
			if filter != "" {
				filter = xs.Slice(filter, 0, xs.Len(filter)-1)
				conf.DbDriver.SetFilter(filter)

				controls.bookTable.SetRowCount(len(conf.DbDriver.FilteredBooks()))
				controls.bookListWindow.SetTitle(fmt.Sprintf("Book list [%s]", filter))
			}
			return true
		case term.KeyEsc:
			go ui.PutEvent(ui.Event{Type: ui.EventCloseWindow})
			return true
		case term.KeyEnter:
			row := controls.bookTable.SelectedRow()
			if row != -1 {
				book := conf.DbDriver.FilteredBooks()[row]
				if book.FilePath != conf.LastFile {
					closeBook(conf)
					loadBook(controls, conf)
				}
			}
			go ui.PutEvent(ui.Event{Type: ui.EventCloseWindow})
			return true
		}
		return false
	}, nil)

	// without overriding this function TableView shows empty values
	controls.bookTable.OnDrawCell(func(info *ui.ColumnDrawInfo) {
		filtered := conf.DbDriver.FilteredBooks()
		if info.Row >= len(filtered) {
			return
		}
		book := filtered[info.Row]
		info.Text = getBookColumnText(book, info.Col)
	})

	// override onSelect to display full cell text in a 'statusbar' - the
	// widget at the bottom of the dialog
	controls.bookTable.OnSelectCell(func(col, row int) {
		if col == -1 || row == -1 {
			return
		}
		filtered := conf.DbDriver.FilteredBooks()
		book := filtered[row]
		controls.bookInfoDetail.SetTitle(getBookColumnText(book, col))
	})

	// override it to do custom sorting and delete a book from library
	controls.bookTable.OnAction(func(ev ui.TableEvent) {
		filtered := conf.DbDriver.FilteredBooks()
		if ev.Action == ui.TableActionDelete {
			if ev.Row == -1 {
				return
			}

			book := filtered[ev.Row]
			controls.bookListWindow.SetModal(false)
			controls.askLabel.SetTitle(fmt.Sprintf("Information about book <c:bright green>'%s'<c:> will be removed from the library. Continue?", book.Title))
			controls.askWindow.SetModal(true)
			controls.askWindow.SetVisible(true)
			ui.ActivateControl(controls.askWindow, controls.askCancel)

			controls.askRemove.OnClick(func(evBtn ui.Event) {
				conf.DbDriver.DeleteBookByIndex(ev.Row)
				controls.bookTable.SetRowCount(controls.bookTable.RowCount() - 1)

				controls.askWindow.SetModal(false)
				controls.askWindow.SetVisible(false)
				ui.ActivateControl(controls.bookListWindow, controls.bookTable)
			})

			return
		}

		if ev.Action != ui.TableActionSort {
			return
		}

		if ev.Col == -1 {
			return
		}
		fields := []string{
			common.FIELD_AUTHOR,
			common.FIELD_TITLE,
			common.FIELD_PERCENT,
			"",
			common.FIELD_GENRE,
			common.FIELD_ADDED,
			common.FIELD_COMPLETED,
		}

		if ev.Sort == ui.SortNone || ev.Col >= len(fields) {
			conf.DbDriver.SetSortMode(common.FIELD_AUTHOR, true)
			return
		}

		field := fields[ev.Col]
		if field == "" {
			conf.DbDriver.SetSortMode(common.FIELD_AUTHOR, ev.Sort == ui.SortAsc)
		} else {
			conf.DbDriver.SetSortMode(field, ev.Sort == ui.SortAsc)
		}
	})
}

// closeBook saves the current reading progress to a database and file
func closeBook(conf *cf.Config) {
	conf.SaveLastFileInfo()

	if conf.UseDb && conf.LastFile != "" && conf.LastLength != 0 {
		var brec common.BookRecord
		brec.FirstName = conf.Info.FirstName
		brec.LastName = conf.Info.LastName
		brec.Title = conf.Info.Title
		brec.Language = conf.Info.Language
		brec.Sequence = conf.Info.Sequence
		brec.Genre = conf.Info.Genre

		conf.DbDriver.UpdateBookInDb(conf.LastFile, conf.LastPosition, conf.LastLength, &brec)
	}
}

// titleForBook generates a short description of a book by its full info
// Used to show it in a reader title
func titleForBook(info fbutils.BookInfo) string {
	var s string

	if info.FirstName != "" {
		r, _ := utf8.DecodeRuneInString(info.FirstName)
		s = string(r) + "." + info.LastName
	} else {
		s = info.LastName
	}
	s += " - " + info.Title

	return s
}

// getFilenameFromArgs looks for a file name in the argument list and
// generates a full path for the book if the path is not absolute
func getFilenameFromArgs(conf *cf.Config) string {
	flag.Parse()

	fileName := flag.Arg(0)
	if fileName != "" && !path.IsAbs(fileName) {
		currDir, _ := os.Getwd()
		if currDir != conf.BinPath {
			absName, err := path.Abs(fileName)
			if err == nil {
				fileName = absName
			}
		}
	}

	return fileName
}

func main() {
	var controls ControlList
	conf := cf.InitConfig()

	// read the last book file name from the configuration file
	// in case of argument list is empty
	fileName := getFilenameFromArgs(conf)
	lastFileConf := conf.LastFile
	if fileName == "" {
		fileName = conf.LastFile
	} else {
		conf.LastFile = fileName
	}

	// read book database if it is ON
	if conf.UseDb {
		conf.InitDatabase()
		conf.DbDriver.ReadDatabase()
	}

	var width int
	ui.InitLibrary()
	defer ui.DeinitLibrary()

	// create reader and format a book to fit the reader width
	createView(&controls, conf)
	createBookConfirm(&controls, conf)
	width, _ = controls.reader.Size()

	var lines []string
    absFileName, _ := path.Abs(fileName)
	conf.Info, lines = fbutils.ParseBook(absFileName, true)
	conf.Lines = fbutils.FormatBook(lines, width, conf.Justify)

	// restore the book position to the latest saved one
	// it is read from the last file info and database
	if conf.LastLength > 0 && conf.LastLength != len(conf.Lines) {
		conf.LastPosition = conf.LastPosition * len(conf.Lines) / conf.LastLength
	}
	if fileName != "" && lastFileConf != conf.LastFile && conf.UseDb {
		b, found := conf.DbDriver.BookByFilePath(fileName)
		if found {
			conf.LastPosition = b.LineLast
			conf.LastLength = b.LineTotal
		} else {
			conf.LastPosition = 0
		}
	}

	savedPos := conf.LastPosition
	controls.reader.SetLineCount(len(conf.Lines))

	// override OnPositionChanged to update the current positon and
	// percent read for an opened book
	controls.reader.OnPositionChanged(func(topLine int, totalLines int) {
		conf.LastPosition = topLine
		conf.LastLength = totalLines

		topLine++
		winTitle := fmt.Sprintf("[%v%%] [%v/%v] %s",
			int(topLine*100/totalLines),
			topLine, totalLines, titleForBook(conf.Info))
		controls.mainWindow.SetTitle(winTitle)
	})

	controls.reader.SetTopLine(savedPos)
	controls.reader.SetLineCount(len(conf.Lines))
	controls.reader.OnDrawLine(func(ind int) string {
		return conf.Lines[ind]
	})

	// start UI loop
	mainLoop(&controls, conf)

	// save the current position and the last book file name on app close
	closeBook(conf)
}
