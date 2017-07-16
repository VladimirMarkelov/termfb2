package common

type BookRecord struct {
	// internal
	FilePath string
	Id       string
	Added     string
	Completed string
	LineLast  int
	LineTotal int
	// from FB2
	FirstName string
	LastName  string
	Title     string
	Sequence  string
	Language  string
	Genre     string
}

type BookDb interface {
	ReadDatabase()
	SetFilter(filter string)
	Filter() string
	FilteredBooks() []BookRecord
	DeleteBookByIndex(index int)
	BookList() []BookRecord
	UpdateBookInDb(bookPath string, position, length int, bookInfo *BookRecord)
	SetSortMode(field string, asc bool)
	BookByFilePath(filePath string) (BookRecord, bool)
}
