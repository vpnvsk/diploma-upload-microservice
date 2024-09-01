package handler

import (
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
	err := h.service.APIClient.GetData(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process data"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Data processed successfully"})
}
