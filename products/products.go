package products

import (
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type Product struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Model string `json:"model"`
}

func newProductFromXLSXRow(cols []string) (Product, error) {
	return Product{
		ID:    cols[0],
		Type:  cols[1],
		Model: cols[2],
	}, nil
}

type DB struct {
	sync.RWMutex
	fileModifyTime time.Time
	m              map[string]Product
}

func New() *DB {
	return &DB{m: make(map[string]Product)}
}

func (c *DB) LoadFromXLSX(xlsx string) error {
	stat, err := os.Stat(xlsx)
	if err != nil {
		return err
	}

	if !c.fileModifyTime.Before(stat.ModTime()) {
		return nil
	}

	f, err := excelize.OpenFile(xlsx)
	if err != nil {
		os.Remove(xlsx)
		logrus.WithField("error", err).Warnf("fail to open local file: %s", xlsx)
		return err
	}
	defer f.Close()

	sheet := f.GetSheetName(0)

	rows, err := f.Rows(sheet)
	if err != nil {
		return err
	}
	defer rows.Close()

	rows.Next() // skip header

	m := make(map[string]Product)
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			logrus.Error("fail to read excel.", err)
			continue
		}

		item, err := newProductFromXLSXRow(row)
		if err != nil {
			logrus.Error("fail to parse row in excel", err)
			continue
		}

		m[item.ID] = item
	}

	c.Lock()
	c.m = m
	c.fileModifyTime = stat.ModTime()
	c.Unlock()

	logrus.Infof("load %d products in to database.", len(c.m))
	return nil
}

func (c *DB) Get(id string) (item Product, isExists bool) {
	c.RLock()
	item, isExists = c.m[id]
	c.RUnlock()
	return
}
