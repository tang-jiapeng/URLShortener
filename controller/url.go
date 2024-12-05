package controller

import (
	"URLShortener/model"
	"URLShortener/service"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type URLHandler struct {
	urlService service.URLService
	baseURL    string
}

func NewURLHandler(urlService service.URLService, baseURL string) *URLHandler {
	if urlService == nil {
		panic("urlService is nil during URLHandler initialization")
	}
	return &URLHandler{
		urlService: urlService,
		baseURL:    baseURL,
	}
}

func (h *URLHandler) CreateURL(c *gin.Context) {
	var req model.CreateURLRequest
	validate := validator.New()
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := validate.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	url, err := h.urlService.CreateURL(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, model.CreateURLResponse{
		ShortURL:  h.baseURL + "/" + url.ShortCode,
		ExpiresAt: url.ExpiresAt,
	})
}

func (h *URLHandler) RedirectURL(c *gin.Context) {
	shortCode := c.Param("code")
	url, err := h.urlService.GetURL(c.Request.Context(), shortCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if url == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "url not found"})
	}
	c.Redirect(http.StatusMovedPermanently, url.OriginalUrl)
}
