package products

import (
	"io"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type Product struct {
	ID string `json:"id"`
}

func newProductFromXLSXRow(cols []string) (Product, error) {
	return Product{ID: cols[0]}, nil
}

type DB struct {
	sync.RWMutex
	m map[string]Product
}

func New() *DB {
	return &DB{m: make(map[string]Product)}
}

func (c *DB) LoadFromXLSX(xlsx io.Reader) error {

	f, err := excelize.OpenReader(xlsx)
	if err != nil {
		return err
	}

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
