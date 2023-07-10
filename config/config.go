package config

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

var (
	WebdavURL    string
	WebdavUser   string
	WebdavPass   string
	WebdavRoot   string
	TempFolder   string
	RegistryFile string
	ProductsFile string
	Listen       string
	LogLevel     int
)

func init() {
	WebdavURL = os.Getenv("NATWIN_WebdavURL")
	WebdavUser = os.Getenv("NATWIN_WebdavUser")
	WebdavPass = os.Getenv("NATWIN_WebdavPass")
	WebdavRoot = os.Getenv("NATWIN_WebdavRoot")
	TempFolder = os.Getenv("NATWIN_LocalFolder")
	RegistryFile = os.Getenv("NATWIN_RegistryFile")
	ProductsFile = os.Getenv("NATWIN_ProductsFile")
	Listen = os.Getenv("NATWIN_Listen")

	if Listen == "" {
		logrus.Fatal("config env not found.")
	}

	lvl, err := strconv.Atoi(os.Getenv("NATWIN_LogLevel"))
	if err != nil {
		logrus.Fatalf("fail to parse log level: %s", err)
	}
	LogLevel = lvl
}
