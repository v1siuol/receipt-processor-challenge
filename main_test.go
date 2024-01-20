package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name           string
	input          []byte
	expectedPoints int64
}

func TestTarget(t *testing.T) {
	server = NewServer()
	router := setupRouter()

	test := testCase{
		name: "Target Receipt",
		input: []byte(`{
			"retailer": "Target",
			"purchaseDate": "2022-01-01",
			"purchaseTime": "13:01",
			"items": [
			  {
				"shortDescription": "Mountain Dew 12PK",
				"price": "6.49"
			  },{
				"shortDescription": "Emils Cheese Pizza",
				"price": "12.25"
			  },{
				"shortDescription": "Knorr Creamy Chicken",
				"price": "1.26"
			  },{
				"shortDescription": "Doritos Nacho Cheese",
				"price": "3.35"
			  },{
				"shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ",
				"price": "12.00"
			  }
			],
			"total": "35.35"
            }`),
		expectedPoints: 28,
	}

	t.Run(test.name, func(t *testing.T) {
		testReceiptProcessing(t, router, test.input, test.expectedPoints)
	})
}

func TestMnM(t *testing.T) {
	server = NewServer()
	router := setupRouter()

	test := testCase{
		name: "M&M Corner Market Receipt",
		input: []byte(`{
			"retailer": "M&M Corner Market",
			"purchaseDate": "2022-03-20",
			"purchaseTime": "14:33",
			"items": [
			  {
				"shortDescription": "Gatorade",
				"price": "2.25"
			  },{
				"shortDescription": "Gatorade",
				"price": "2.25"
			  },{
				"shortDescription": "Gatorade",
				"price": "2.25"
			  },{
				"shortDescription": "Gatorade",
				"price": "2.25"
			  }
			],
			"total": "9.00"
            }`),
		expectedPoints: 109,
	}

	t.Run(test.name, func(t *testing.T) {
		testReceiptProcessing(t, router, test.input, test.expectedPoints)
	})
}

func TestMorning(t *testing.T) {
	server = NewServer()
	router := setupRouter()

	test := testCase{
		name: "morning-receipt",
		input: []byte(`{
			"retailer": "Walgreens",
			"purchaseDate": "2022-01-02",
			"purchaseTime": "08:13",
			"total": "2.65",
			"items": [
				{"shortDescription": "Pepsi - 12-oz", "price": "1.25"},
				{"shortDescription": "Dasani", "price": "1.40"}
			]
		}`),
		expectedPoints: 15,
	}

	t.Run(test.name, func(t *testing.T) {
		testReceiptProcessing(t, router, test.input, test.expectedPoints)
	})
}

func TestSimple(t *testing.T) {
	server = NewServer()
	router := setupRouter()

	test := testCase{
		name: "simple-receipt",
		input: []byte(`{
			"retailer": "Target",
			"purchaseDate": "2022-01-02",
			"purchaseTime": "13:13",
			"total": "1.25",
			"items": [
				{"shortDescription": "Pepsi - 12-oz", "price": "1.25"}
			]
		}`),
		expectedPoints: 31,
	}

	t.Run(test.name, func(t *testing.T) {
		testReceiptProcessing(t, router, test.input, test.expectedPoints)
	})
}

func testReceiptProcessing(t testing.TB, router *gin.Engine, input []byte, expectedPoints int64) {
	// 1. Submit receipt
	req, _ := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(input))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var postResponse map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &postResponse)
	assert.Nil(t, err)
	assert.True(t, validateJSONField(postResponse, "id"))

	// 2. Get the points
	id := postResponse["id"]
	req, _ = http.NewRequest("GET", "/receipts/"+id+"/points", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req)

	assert.Equal(t, http.StatusOK, w2.Code)
	var getResponse map[string]int64
	err = json.Unmarshal(w2.Body.Bytes(), &getResponse)
	assert.Nil(t, err)
	points := getResponse["points"]
	assert.Equal(t, expectedPoints, points)
}

func validateJSONField(data map[string]string, key string) bool {
	if _, exists := data[key]; !exists {
		return false
	}
	notEmpty, _ := regexp.MatchString(`^\S+$`, data[key])
	return notEmpty
}
