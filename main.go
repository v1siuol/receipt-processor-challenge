package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

var server *Server

func setupRouter() *gin.Engine {
	router := gin.Default()

	router.POST("/receipt/process", receiptPOST)
	router.GET("/receipt/:id/points", receiptGET)

	return router
}

// Takes in a JSON receipt and returns a JSON object with an ID generated.
func receiptPOST(c *gin.Context) {
	var receipt Receipt
	if err := c.BindJSON(&receipt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The receipt is invalid"})
		return
	}

	id, err := server.Submit(receipt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The receipt is invalid"})
		return
	}

	c.JSON(http.StatusOK, ReceiptResponse{ID: id})
}

func receiptGET(c *gin.Context) {
	id := c.Param("id")
	if points, exists := server.GetPoints(id); exists {
		c.JSON(http.StatusOK, PointsResponse{Points: points})
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "No receipt found for that id"})
}

func main() {
	server = NewServer()
	router := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	// TODO support port for config
	// TODO docs
	// TODO GitHub Actions
	// TODO tests
	router.Run(":8080")
}
