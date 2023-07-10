package server

import (
	"fmt"
	"natwin/registry"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	ProductNotFound     = "ProductNotFound"
	ProductRegistered   = "ProductRegistered"
	FormMissingRequired = "FormMissingRequired"
	FormValidationFail  = "FormValidationFail"
)

var dav *Webdav

func getWebdav() *Webdav {
	if dav != nil {
		return dav
	}

	dav = NewWebdav()
	return dav
}

func getProduct(c *gin.Context) {

	productID := c.Param("productID")

	p, err := getWebdav().Products()
	if err != nil {
		logrus.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "server internal error",
		})
		return
	}

	product, flg := p.Get(productID)
	if !flg {
		c.JSON(http.StatusNotFound, gin.H{
			"Code": ProductNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, product)
}

func getRegisterPage(c *gin.Context) {

	productID := c.Param("productID")
	p, err := getWebdav().Products()
	if err != nil {
		logrus.WithField("productID", productID).Errorf("fail to get product info: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{})
		return
	}

	product, isExists := p.Get(productID)
	if !isExists {
		c.HTML(http.StatusNotFound, "fail.html", gin.H{
			"Code":      ProductNotFound,
			"ProductID": productID,
		})
		return
	}

	c.HTML(http.StatusOK, "register.html", product)
}

var (
	regChineseName  = regexp.MustCompile(`^[\p{Han}]+$`)
	regCellphone    = regexp.MustCompile(`^1\d{10}$`)
	regEmail        = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	regPurchaseDate = regexp.MustCompile(`^\d{4}\-\d{2}\-\d{2}$`)
)

type RegistrationForm struct {
	ProductID    string `form:"product_id" binding:"required"`
	Forename     string `form:"forename"`
	Surname      string `form:"surname" binding:"required"`
	Title        string `form:"title" binding:"required"`
	Phone        string `form:"phone" binding:"required"`
	Email        string `form:"email" binding:"required"`
	PurchaseDate string `form:"purchase_date" binding:"required"`
	PurchaseFrom string `form:"buy_from" binding:"required"`
}

func (f *RegistrationForm) Validate() error {
	if !regChineseName.MatchString(f.Surname) {
		return fmt.Errorf("incorrect chinese surname: %s", f.Surname)
	}

	if f.Forename != "" && !regChineseName.MatchString(f.Forename) {
		return fmt.Errorf("incorrect chinese forename: %s", f.Forename)
	}

	if f.Title != "Mr" && f.Title != "Ms" {
		return fmt.Errorf("incorrect title: %s", f.Title)
	}

	if !regCellphone.MatchString(f.Phone) {
		return fmt.Errorf("incorrect phone number: %s", f.Phone)
	}

	if !regEmail.MatchString(f.Email) {
		return fmt.Errorf("incorrect email: %s", f.Email)
	}

	if !regPurchaseDate.MatchString(f.PurchaseDate) {
		return fmt.Errorf("incorrect purchase date: %s", f.PurchaseDate)
	}

	return nil
}

func (f RegistrationForm) ToRegistration() registry.Registration {
	return registry.Registration{
		ProductID:    f.ProductID,
		Forename:     f.Forename,
		Surname:      f.Surname,
		Title:        f.Title,
		Phone:        f.Phone,
		Email:        f.Email,
		PurchaseDate: f.PurchaseDate,
		PurchaseFrom: f.PurchaseFrom,
		RegisterTime: time.Now(),
	}
}

func registerProduct(c *gin.Context) {
	var form RegistrationForm
	if err := c.ShouldBind(&form); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"ip": c.ClientIP(),
			"ua": c.GetHeader("User-Agent"),
		}).Warn("incorrect form submitted.")
		
		c.HTML(http.StatusPreconditionFailed, "fail.html", gin.H{
			"Code": FormMissingRequired,
			"ProductID": "Unknown",
		})
		return
	}

	if err := form.Validate(); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"ip": c.ClientIP(),
			"ua": c.GetHeader("User-Agent"),
		}).Warn("incorrect form submitted.")
		
		c.HTML(http.StatusPreconditionFailed, "fail.html", gin.H{
			"Code": FormValidationFail,
			"ProductID": form.ProductID,
		})
	}

	webdav := getWebdav()

	p, err := webdav.Products()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{})
		return
	}

	if _, flg := p.Get(form.ProductID); !flg {
		c.HTML(http.StatusNotFound, "fail.html", gin.H{
			"Code":      ProductNotFound,
			"ProductID": form.ProductID,
		})
		return
	}

	r, err := webdav.Registration()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{})
		return
	}

	if err := r.Register(form.ProductID); err != nil {
		c.HTML(http.StatusPreconditionFailed, "fail.html", gin.H{
			"Code":      ProductRegistered,
			"ProductID": form.ProductID,
		})
		return
	}

	if err := webdav.WriteRegistration(form.ToRegistration()); err != nil {
		logrus.Errorf("fail to write registration: %s", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{})
		return
	}

	c.HTML(http.StatusOK, "success.html", gin.H{
		"ProductID":   form.ProductID,
		"UserSurname": form.Surname,
		"UserTitle":   form.Title,
	})
}

func AddRouter(router gin.IRouter) {
	router.GET("/products/:productID", getProduct)
	router.GET("/registry/:productID", getRegisterPage)
	router.POST("/registry", registerProduct)
}
