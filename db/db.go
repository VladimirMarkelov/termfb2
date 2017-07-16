package db

import (
	"encoding/json"
	"fmt"
	"github.com/VladimirMarkelov/termfb2/common"
	scribble "github.com/nanobox-io/golang-scribble"
	"github.com/nu7hatch/gouuid"
	"os"
	path "path/filepath"
	"sort"
	"strings"
	"time"
)

type ScribbleDb struct {
	dbDriver *scribble.Driver

	bookMap      map[string]common.BookRecord
	bookList     []common.BookRecord
	bookFiltered []common.BookRecord

	filter   string
	sortMode string
	sortAsc  bool
}

func InitDb(dbPath string) *ScribbleDb {
	db := new(ScribbleDb)
	db.sortMode = common.FIELD_AUTHOR
	db.sortAsc = true
	db.dbDriver, _ = scribble.New(path.Join(dbPath, common.DBFILE), nil)

	db.ReadDatabase()

	return db
}

func (db *ScribbleDb) ReadDatabase() {
	records, err := db.dbDriver.ReadAll(common.DBCOLLECTION)
	if err != nil {
		return
	}

	db.bookMap = make(map[string]common.BookRecord, 0)
	db.bookList = make([]common.BookRecord, 0)
	db.bookFiltered = make([]common.BookRecord, 0)
	for _, r := range records {
		b := common.BookRecord{}
		if err := json.Unmarshal([]byte(r), &b); err != nil {
			fmt.Printf("Failed to parse book list: %v\n", err)
			os.Exit(1)
		}
		db.bookMap[b.FilePath] = b
		db.bookList = append(db.bookList, b)
		db.bookFiltered = append(db.bookFiltered, b)
	}
	db.bookArraySort()
}

func (db *ScribbleDb) UpdateBookInDb(bookPath string, pos, length int, bookInfo *common.BookRecord) {
	book, found := db.bookMap[bookPath]
	if found {
		if book.LineLast != pos || book.LineTotal != length {
			book.LineLast = pos
			book.LineTotal = length
			if pos+1 == length && book.Completed == "" {
				t := time.Now()
				book.Completed = t.Format(time.RFC3339)
			}
			db.dbDriver.Write(common.DBCOLLECTION, book.Id, &book)
		}
	} else {
		var book common.BookRecord
		uid, _ := uuid.NewV4()
		book.Id = uid.String()
		book.Title = bookInfo.Title
		book.FirstName = bookInfo.FirstName
		book.LastName = bookInfo.LastName
		book.Sequence = bookInfo.Sequence
		book.Language = bookInfo.Language
		book.Genre = bookInfo.Genre
		book.FilePath = bookPath
		t := time.Now()
		book.Added = t.Format(time.RFC3339)
		book.LineLast = pos
		book.LineTotal = length
		if db.bookMap == nil {
			db.bookMap = make(map[string]common.BookRecord, 0)
		}
		db.bookMap[book.FilePath] = book
		db.dbDriver.Write(common.DBCOLLECTION, book.Id, &book)
	}
}

func (db *ScribbleDb) bookMapToArray() []common.BookRecord {
	arr := make([]common.BookRecord, 0, len(db.bookMap))
	for _, b := range db.bookMap {
		arr = append(arr, b)
	}

	return arr
}

func (db *ScribbleDb) bookFilter() []common.BookRecord {
	list := make([]common.BookRecord, 0)

	if len(db.bookList) == 0 {
		return list
	}

	flt := strings.ToLower(db.filter)
	for _, b := range db.bookList {
		if db.filter == "" {
			list = append(list, b)
		} else {
			if strings.Index(strings.ToLower(b.FirstName), flt) != -1 ||
				strings.Index(strings.ToLower(b.LastName), flt) != -1 ||
				strings.Index(strings.ToLower(b.Title), flt) != -1 ||
				strings.Index(strings.ToLower(b.FilePath), flt) != -1 ||
				strings.Index(strings.ToLower(b.Sequence), flt) != -1 {
				list = append(list, b)
			}
		}
	}

	return list
}

func (db *ScribbleDb) compareByAuthorTitleSequence(b1 *common.BookRecord, b2 *common.BookRecord, asc bool) bool {
	if b1.LastName < b2.LastName {
		return asc
	} else if b1.LastName > b2.LastName {
		return !asc
	} else if b1.FirstName < b2.FirstName {
		return asc
	} else if b1.FirstName > b2.FirstName {
		return !asc
	} else if b1.Title < b2.Title {
		return asc
	} else if b1.Title > b2.Title {
		return !asc
	} else {
		if (b1.Sequence < b2.Sequence && asc) ||
			(b1.Sequence > b2.Sequence && !asc) {
			return true
		} else {
			return false
		}
	}
}

func (db *ScribbleDb) bookArraySort() {
	switch db.sortMode {
	case common.FIELD_GENRE:
		sort.SliceStable(db.bookFiltered, func(i, j int) bool {
			if db.bookFiltered[i].Genre < db.bookFiltered[j].Genre {
				return db.sortAsc
			} else if db.bookFiltered[i].Genre > db.bookFiltered[j].Genre {
				return !db.sortAsc
			} else {
				return db.compareByAuthorTitleSequence(&db.bookFiltered[i], &db.bookFiltered[j], db.sortAsc)
			}
		})
	case common.FIELD_TITLE:
		sort.SliceStable(db.bookFiltered, func(i, j int) bool {
			if db.bookFiltered[i].Title < db.bookFiltered[j].Title {
				return db.sortAsc
			} else if db.bookFiltered[i].Title > db.bookFiltered[j].Title {
				return !db.sortAsc
			} else {
				return db.compareByAuthorTitleSequence(&db.bookFiltered[i], &db.bookFiltered[j], db.sortAsc)
			}
		})
	case common.FIELD_ADDED:
		sort.SliceStable(db.bookFiltered, func(i, j int) bool {
			if db.bookFiltered[i].Added < db.bookFiltered[j].Added {
				return db.sortAsc
			} else if db.bookFiltered[i].Added > db.bookFiltered[j].Added {
				return !db.sortAsc
			} else {
				return db.compareByAuthorTitleSequence(&db.bookFiltered[i], &db.bookFiltered[j], db.sortAsc)
			}
		})
	case common.FIELD_COMPLETED:
		sort.SliceStable(db.bookFiltered, func(i, j int) bool {
			if db.bookFiltered[i].Completed < db.bookFiltered[j].Completed {
				return db.sortAsc
			} else if db.bookFiltered[i].Completed > db.bookFiltered[j].Completed {
				return !db.sortAsc
			} else {
				return db.compareByAuthorTitleSequence(&db.bookFiltered[i], &db.bookFiltered[j], db.sortAsc)
			}
		})
	case common.FIELD_PERCENT:
		sort.SliceStable(db.bookFiltered, func(i, j int) bool {
			prc1 := 0
			if db.bookFiltered[i].LineTotal != 0 {
				prc1 = db.bookFiltered[i].LineLast * 100 / db.bookFiltered[i].LineTotal
			}
			prc2 := 0
			if db.bookFiltered[j].LineTotal != 0 {
				prc2 = db.bookFiltered[j].LineLast * 100 / db.bookFiltered[j].LineTotal
			}

			if prc1 < prc2 {
				return db.sortAsc
			} else if prc1 > prc2 {
				return !db.sortAsc
			} else {
				return db.compareByAuthorTitleSequence(&db.bookFiltered[i], &db.bookFiltered[j], db.sortAsc)
			}
		})
	case common.FIELD_AUTHOR:
		sort.SliceStable(db.bookFiltered, func(i, j int) bool {
			return db.compareByAuthorTitleSequence(&db.bookFiltered[i], &db.bookFiltered[j], db.sortAsc)
		})
	default:
		// at this moment default equals FIELD_AUTHOR
		sort.SliceStable(db.bookFiltered, func(i, j int) bool {
			return db.compareByAuthorTitleSequence(&db.bookFiltered[i], &db.bookFiltered[j], db.sortAsc)
		})
	}
}

func (db *ScribbleDb) SetFilter(filter string) {
	if filter == db.filter {
		return
	}

	db.filter = filter
	db.bookFiltered = db.bookFilter()
	db.bookArraySort()
}

func (db *ScribbleDb) Filter() string {
	return db.filter
}

func (db *ScribbleDb) FilteredBooks() []common.BookRecord {
	return db.bookFiltered
}

func (db *ScribbleDb) DeleteBookByIndex(index int) {
	if index < 0 || index >= len(db.bookFiltered) {
		return
	}

	book := db.bookFiltered[index]
	db.dbDriver.Delete(common.DBCOLLECTION, book.Id)
	if index == len(db.bookFiltered)-1 {
		db.bookFiltered = db.bookFiltered[:index]
	} else {
		db.bookFiltered = append(db.bookFiltered[:index], db.bookFiltered[index+1:]...)
	}

	ind := -1
	for i, b := range db.bookList {
		if b.Id == book.Id {
			ind = i
			break
		}
	}
	if ind != -1 {
		if ind == len(db.bookList)-1 {
			db.bookList = db.bookList[:ind]
		} else {
			db.bookList = append(db.bookList[:ind], db.bookList[ind+1:]...)
		}
	}
}

func (db *ScribbleDb) BookList() []common.BookRecord {
	return db.bookList
}

func (db *ScribbleDb) SetSortMode(field string, asc bool) {
	if field != db.sortMode || asc != db.sortAsc {
		db.sortMode = field
		db.sortAsc = asc

		if len(db.bookFiltered) > 0 {
			db.bookArraySort()
		}
	}
}

func (db *ScribbleDb) BookByFilePath(filePath string) (common.BookRecord, bool) {
	b, found := db.bookMap[filePath]
	return b, found
}
