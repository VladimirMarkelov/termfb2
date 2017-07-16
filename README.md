# Termfb2
It a simple FB2 reader for terminals. The reader does not support all FB2 features but it allows to quickly open and read any FB2 file without running graphical DE - all you need is terminal(in Linux) or cmd(in Windows).
A demo of termfb in action (the application is in portable mode with changed text and background colors):

<img src="./images/termfb2_in_action.gif" alt="TermFB2 in action">


# Features and limitations
## Features
* Does not requires any external libraries or GUI to open FB2 file
* The application detects if FB2 file is zipped and unpacks it automatically before reading
* Remembers last opened file and position in it (it works always and does not depend on library)
* Optional (enabled by default) library - a book is added to the library automatically after opening the book. The library stores the following information about every book: author, title, sequence, genre, language, date added, date completed, the last saved position in the book (so you can read a few book in turns and continue every time from the line you stopped the last time), file path(if the book is somewhere in the directory or sub-directory where executable file is then the path is relative and absolute otherwise - it helps to create a portable installation)
* The library has simple lookup: incremental filter. Just start typing inside the library and the book list is automatically filtered. You do not need to choose what column to use for filtering - the application looks for the entered text at the same time in columns author, title, sequence, and file path
* The reader does not have settings inside the application but there is a manually editable configuration file (please see termfb2.conf.example as an example). The application reads it at start but never writes anything to it. So you can edit it as you wish and all changes are kept. Configuration file syntax is very simple: lines that starts with # is a comment line, otherwise it must be in **key=value** format
* The reader is not portable by default and writes database and reads configuration from "user home directory"/.rionnag/termfb2. But you can convert it to portable version by creating a configuration file (it can be empty file) termfb2.conf in the same directory where the executable is before launching the reader
* Two ways of displaying the text: with and without justification. Examples of how both modes look like, please, see images here: ![text justification](https://github.com/VladimirMarkelov/fb2text)
* When the text is scrolled by page up/down then the last/first visible line is kept to make reading more comfortable
* In the library columns show short text but if you select any cell then in the statusbar you can see the full value of the column

## Limitations
* Only UTF8 FB2 books are supported
* No colors (besides text color and background color that can be changed by editing termfb.com). The reader understands FB2 tags like **strong** and **emphasis** but does not do anything in this version
* Footnotes are not supported yet. The reader stops reading FB2 file at the end of the first **body** section
* No search available in reading mode yet
* Dynamic console resizing is not fully supported: application windows use the new terminal size but if a book is already opened its text is not reformatted to use the new width. So the book must be reopened to fit it to the new terminal width
* Terminal size should be at least 30 lines height (minimal width around 50-60 columns)

# Application arguments
At this moment the reader does not have any command line arguments except file name. If you start the reader without arguments then it reads the last opened book information and opens that book. If you provide a file name then the application do the following: at first it checks if it is the same file that was opened the last - in this case it restores the position from the last info file, if it is not the last opened book then the application looks for the book in the library and tried to retrieve position information from the database. If both ways fail then the reader opens the book from the beginning.

# Hotkeys
## Global hotkeys
* CtrlQ + CtrlQ - close application
## Reader mode
* Arrow Down or J - scrolls line down
* Arrow Up or K - scrolls line up
* Page Up or U - scrolls page up
* Space or Page Down or D - scrolls page down
* F2 - opens the book library (if it is enabled)
## Library dialog
* Escape - closes the library
* Enter - opens the selected book
* F4 - sorts the book list by the selected column (multiple pressing the key changes the mode in a cycle: ascending, descending, off - column marker in column header shows the current mode). If sort mode is off then the default sorting is used: by author, title and sequence
* Any printable character - incremental filter, the current filter is displayed in dialog title
* Backspace - erase the last filter letter if filter is not empty
* Delete - after you confirm the action (choose a button with TAB key, by default **Cancel** button is selected) delete information about selected book from the library (the file is not deleted)

# Troubleshooting
* Book does not open - the reader shows only '--- THE END ---'. Check if the FB2 file is in UTF-8 encoding(I got that issue on files with 'windows-1252' encoding)
* The reader opens book but does not show any text (yet in the title the number of lines and book title are correct) or library opens but does not display book list - check if the size of terminal window is at least 30 lines height(it is enough for reader, 40 lines is enough to fix the library dialog) and 50-60 column width

# Files used and created by the application
Note: if application is in portable mode then all files are created inside the directory where the executable is. Otherwise all directories and files are created in a user home directory.
* sub-directory ".rionnag" - the application keeps everything inside it
* file **.rionnag/last** - name of the last opened book and position in it
* sub-directory **.rionnag/book.db/books/** - a book database, one book - one file. The directory is created only if a database is enabled (see information about configuration file below)
* optional file that does not exist by default (use termfb2.conf.example as an example file) **.rionnag/termfb2.conf** - configuration file. The application only reads it and never writes to it. At this moment there are 4 options available:
- **useDb** - use database to keep information about all read books. It is enabled by default(useDb=1), disable it by setting useDb to 0
- **textColor** - a color of text in the reader (library dialog is not affected by this option). Default value is 'default' that means 'use color that is default for the current theme ". Available colors are: black, yellow, red, green, blue, magenta, cyan, and white. And you can intensify color by adding 'bold' or 'bright' to color (before or after color name). Examples of correct colors: "textColor=red", "textColor=while bright", "textColor="bold red", "textColor=green+bright"
- **backColor** - a color of background in the reader. Please read details in **textColor** section
- **justify** - display justified or uneven lines. Default value is 0 - justification is disabled
