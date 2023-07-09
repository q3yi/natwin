package server

import (
	"fmt"
	"natwin/registry"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
			"message": "product id not exists",
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
			"message": "product not exists",
		})
		return
	}

	c.HTML(http.StatusOK, "register.html", product)
}

type RegistrationForm struct {
	ProductID string `form:"product_id" binding:"required"`
	Forename  string `form:"forename"`
	Surname   string `form:"surname" binding:"required"`
	Phone     string `form:"phone" binding:"required"`
}

func (f RegistrationForm) ToRegistration() registry.Registration {
	return registry.Registration{
		ProductID: f.ProductID,
		Forename:  f.Forename,
		Surname:   f.Surname,
		Phone:     f.Phone,
	}
}

func registerProduct(c *gin.Context) {
	var form RegistrationForm
	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusPreconditionFailed, "fail.html", gin.H{
			"message": err.Error(),
		})
		return
	}

	webdav := getWebdav()

	p, err := webdav.Products()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{})
		return
	}

	if _, flg := p.Get(form.ProductID); !flg {
		c.HTML(http.StatusNotFound, "fail.html", gin.H{
			"message": "incorrect product id",
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
			"message": err.Error(),
		})
		return
	}

	if err := webdav.WriteRegistration(form.ToRegistration()); err != nil {
		logrus.Errorf("fail to write registration: %s", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{})
		return
	}

	c.HTML(http.StatusOK, "success.html", gin.H{
		"productId": form.ProductID,
		"username":  fmt.Sprintf("%s%s", form.Surname, form.Forename),
	})
}

func AddRouter(router gin.IRouter) {
	router.GET("/products/:productID", getProduct)
	router.GET("/registry/:productID", getRegisterPage)
	router.POST("/registry", registerProduct)
}
