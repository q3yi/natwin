package registry

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
)

type Registration struct {
	ProductID     string
	Forename      string
	Surname       string
	Gender        string
	Phone         string
	Email         string
	Region        string
	PurchaseTime  string
	ReceiptNumber string
	ProductIDPic  string
	ReceiptPic    string
	ProductPic    string
	RegisterTime  time.Time
}

type DB struct {
	registered *sync.Map
}

func New() *DB {
	r := &DB{registered: &sync.Map{}}
	return r
}

func (r *DB) LoadRegisteredFromXLSX(xlsx io.Reader) error {
	f, err := excelize.OpenReader(xlsx)
	if err != nil {
		return err
	}

	sheet1 := f.GetSheetName(0)

	rows, err := f.Rows(sheet1)
	if err != nil {
		return err
	}
	defer rows.Close()

	rows.Next() // skip header

	total := 0
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			continue
		}

		total++
		r.registered.LoadOrStore(row[0], true)
	}

	logrus.Infof("loads %d registration from xlsx.", total)
	return nil
}

func (r *DB) Register(id string) error {
	_, loaded := r.registered.LoadOrStore(id, true)
	if loaded {
		return fmt.Errorf("product(%s) already registered", id)
	}
	return nil
}

func (r *DB) Dergister(id string) {
	r.registered.Delete(id)
}

func (r *DB) WriteToXLSX(u Registration, xlsx string) error {
	f, err := excelize.OpenFile(xlsx)
	if err != nil {
		return err
	}

	sheet1 := f.GetSheetName(0)

	if err := f.InsertRows(sheet1, 2, 1); err != nil {
		return err
	}

	f.SetCellStr(sheet1, "A2", u.ProductID)
	f.SetCellStr(sheet1, "B2", u.Surname)
	f.SetCellStr(sheet1, "C2", u.Phone)

	if err := f.Save(); err != nil {
		return err
	}

	logrus.Infof("write registration to xlsx, id: %s", u.ProductID)
	return nil
}
