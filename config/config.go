package config

import (
	"bufio"
	"fmt"
	ui "github.com/VladimirMarkelov/clui"
	fbutils "github.com/VladimirMarkelov/fb2text"
	"github.com/VladimirMarkelov/termfb2/common"
	"github.com/VladimirMarkelov/termfb2/db"
	homedir "github.com/mitchellh/go-homedir"
	term "github.com/nsf/termbox-go"
	"os"
	path "path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	confPath string

	BinPath   string
	Portable  bool
	BackColor term.Attribute
	TextColor term.Attribute
	Justify   bool

	// info about last opened book
	// lastPosition and lastLength are used in case of DB is off
	LastFile     string
	LastPosition int
	LastLength   int
	LastError    int

	Lines []string

	UseDb    bool
	DbDriver common.BookDb

	Info fbutils.BookInfo
}

func InitConfig() *Config {
	conf := new(Config)

	conf.detectPaths()
	conf.readOptions()

	return conf
}

func (conf *Config) detectPaths() {
	hd, homeerr := homedir.Dir()
	if homeerr != nil {
		hd, homeerr = os.Getwd()
		if homeerr != nil {
			panic("Failed to get both user and current directory")
		}
	}

	execName, _ := os.Executable()
	conf.BinPath = path.Dir(execName)

	cfile := path.Join(conf.BinPath, common.CONFIGFILE)
	if _, err := os.Stat(cfile); os.IsNotExist(err) {
		conf.confPath = path.Join(hd, common.VENDOR, common.APPNAME)
		conf.Portable = false
	} else {
		conf.confPath = conf.BinPath
		conf.Portable = true
	}
}

func (conf *Config) ReadLastFileInfo() {
	conf.LastFile = ""
	conf.LastPosition = 0
	conf.LastLength = 0

	file, err := os.Open(path.Join(conf.confPath, common.LASTFILE))
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	ok := scanner.Scan()
	if ok {
		conf.LastFile = strings.TrimSpace(scanner.Text())
	}
	ok = scanner.Scan()
	if ok {
		if n, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil {
			conf.LastPosition = n
		} else {
			conf.LastPosition = 0
		}
	}
	ok = scanner.Scan()
	if ok {
		if n, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil {
			conf.LastLength = n
		} else {
			conf.LastLength = 0
		}
	}
}

func (conf *Config) SaveLastFileInfo() {
	if conf.LastFile == "" || conf.LastLength == 0 {
		return
	}

	os.MkdirAll(conf.confPath, os.ModeDir|0777)

	file, err := os.Create(path.Join(conf.confPath, common.LASTFILE))
	if err != nil {
		return
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("%v\n", conf.LastFile))
	file.WriteString(fmt.Sprintf("%v\n", conf.LastPosition))
	file.WriteString(fmt.Sprintf("%v\n", conf.LastLength))
}

func (conf *Config) readOptions() {
	conf.BackColor = term.ColorDefault
	conf.TextColor = term.ColorDefault
	conf.UseDb = true
	conf.LastFile = ""

	file, err := os.Open(path.Join(conf.confPath, common.CONFIGFILE))
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "/") {
			continue
		}

		items := strings.SplitN(line, "=", 2)
		if len(items) != 2 {
			continue
		}

		name := strings.TrimSpace(items[0])
		value := strings.TrimSpace(items[1])

		if strings.EqualFold(name, "useDb") {
			conf.UseDb = (value == "1" || strings.EqualFold(value, "on") || strings.EqualFold(value, "true"))
		} else if strings.EqualFold(name, "textColor") {
			conf.TextColor = ui.StringToColor(value)
		} else if strings.EqualFold(name, "backColor") {
			conf.BackColor = ui.StringToColor(value)
		} else if strings.EqualFold(name, "justify") {
			conf.Justify = (value == "1" || strings.EqualFold(value, "on") || strings.EqualFold(value, "true"))
		}
	}
}

func (conf *Config) InitDatabase() {
	if !conf.UseDb {
		return
	}

	conf.DbDriver = db.InitDb(conf.confPath)
}
