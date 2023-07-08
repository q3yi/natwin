package server

import (
	"io"
	"natwin/config"
	"natwin/products"
	"natwin/registry"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/studio-b12/gowebdav"
)

type client struct {
	URL  string
	User string
	Pass string
}

func (c client) Connect() (*gowebdav.Client, error) {
	cli := gowebdav.NewClient(c.URL, c.User, c.Pass)
	if err := cli.Connect(); err != nil {
		return nil, err
	}

	return cli, nil
}

type Webdav struct {
	client           client
	reg              *registry.DB
	registryXLSX     string
	prod             *products.DB
	prodXLSX         string
	localRoot        string
	remoteRoot       string
	cacheLock        *sync.Mutex
	cacheFileModTime map[string]time.Time
}

func NewWebdav() *Webdav {
	cli := client{
		URL:  config.WebdavURL,
		User: config.WebdavUser,
		Pass: config.WebdavPass,
	}
	return &Webdav{
		client:           cli,
		registryXLSX:     config.RegistryFile,
		prodXLSX:         config.ProductsFile,
		localRoot:        config.TempFolder,
		remoteRoot:       config.WebdavRoot,
		cacheLock:        &sync.Mutex{},
		cacheFileModTime: make(map[string]time.Time),
	}
}

func (w *Webdav) filepath(file string) (localFile, remoteFile string) {
	localFile = filepath.Join(w.localRoot, file)
	remoteFile = filepath.Join(w.remoteRoot, file)
	return
}

func (w *Webdav) updateFileCache(localFile, remoteFile string) (updated bool, err error) {
	cli, err := w.client.Connect()
	if err != nil {
		return
	}

	stat, err := cli.Stat(remoteFile)
	if err != nil {
		return
	}

	w.cacheLock.Lock()

	oldModTime := w.cacheFileModTime[remoteFile]
	if oldModTime.Compare(stat.ModTime()) >= 0 {
		w.cacheLock.Unlock()
		return false, nil
	}

	defer w.cacheLock.Unlock()

	reader, err := cli.ReadStream(remoteFile)
	if err != nil {
		return
	}

	dir := filepath.Dir(localFile)
	if err = os.MkdirAll(dir, 0755); err != nil {
		return
	}

	f, err := os.OpenFile(localFile, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return
	}
	defer f.Close()

	if _, err = io.Copy(f, reader); err != nil {
		return
	}

	w.cacheFileModTime[remoteFile] = stat.ModTime()

	logrus.Debugf("file updated: %s, from %s to %s",
		remoteFile,
		oldModTime.Format(time.DateTime),
		stat.ModTime().Format(time.DateTime),
	)

	return true, nil
}

func (w *Webdav) readFile(localFile string, handler func(io.Reader) error) error {
	reader, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer reader.Close()

	return handler(reader)
}

func (w *Webdav) Products() (*products.DB, error) {
	if w.prod == nil {
		w.prod = products.New()
	}

	// try loading from remote file every time a new request coming
	localFile, remoteFile := w.filepath(w.prodXLSX)
	updated, err := w.updateFileCache(localFile, remoteFile)
	if err != nil {
		return nil, err
	}

	if !updated {
		return w.prod, nil
	}

	if err := w.readFile(localFile, w.prod.LoadFromXLSX); err != nil {
		return nil, err
	}

	return w.prod, nil
}

func (w *Webdav) Registration() (*registry.DB, error) {
	if w.reg != nil {
		return w.reg, nil
	}

	// load from remote only once before registory initialized
	localFile, remoteFile := w.filepath(w.registryXLSX)
	_, err := w.updateFileCache(localFile, remoteFile)
	if err != nil {
		return nil, err
	}

	reg := registry.New()
	if err := w.readFile(localFile, reg.LoadRegisteredFromXLSX); err != nil {
		return nil, err
	}

	w.reg = reg
	return w.reg, nil
}

func (w *Webdav) WriteRegistration(reg registry.Registration) error {
	r, err := w.Registration()
	if err != nil {
		return err
	}

	localFile, _ := w.filepath(w.registryXLSX)
	if err := r.WriteToXLSX(reg, localFile); err != nil {
		return err
	}

	return w.writeRegistryToRemote()
}

func (w *Webdav) writeRegistryToRemote() error {
	local, remote := w.filepath(w.registryXLSX)

	cli, err := w.client.Connect()
	if err != nil {
		return err
	}

	f, err := os.Open(local)
	if err != nil {
		return err
	}
	defer f.Close()

	return cli.WriteStream(remote, f, 0755)
}
