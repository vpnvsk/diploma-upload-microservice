package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"net/http"
)

func (h *Handler) UploadPatents(c *gin.Context) {
	var input model.UploadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	err := h.service.APIClientInterface.GetData(input)
	fmt.Println("after req")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process data"})
		return
	}
	fmt.Println("after all")

	c.JSON(http.StatusOK, gin.H{"message": "Data processed successfully"})
}
