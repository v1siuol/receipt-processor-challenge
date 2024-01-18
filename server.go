package main

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Receipt struct {
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type ReceiptResponse struct {
	ID string `json:"id"`
}

type PointsResponse struct {
	Points int64 `json:"points"`
}

type Server struct {
	lock          sync.RWMutex     // global RW Mutex lock, n readers or one writer
	receiptPoints map[string]int64 // stored in-memory mapping of receipt IDs to points
}

// input port?
func NewServer() *Server {
	server := &Server{
		receiptPoints: make(map[string]int64),
	}

	return server
}

func (s *Server) Submit(receipt Receipt) (string, error) {
	if err := validateReceipt(receipt); err != nil {
		return "", err
	}
	var id string
	var points = calcPoints(receipt)
	for attempts := 0; attempts < 3; attempts++ {
		id = generateID()

		s.lock.Lock()
		if _, exists := s.receiptPoints[id]; !exists {
			s.receiptPoints[id] = points
			s.lock.Unlock()
			return id, nil // Unique ID found, return with success
		}
		s.lock.Unlock()

		// Retry if ID was a duplicate
		time.Sleep(time.Millisecond * 100)
	}

	return "", errors.New("failed to generate a unique ID after 3 attempts")
}

func (s *Server) GetPoints(id string) (int64, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	points, exists := s.receiptPoints[id]
	return points, exists
}

func generateID() string {
	return uuid.New().String()
}

func calcPoints(receipt Receipt) int64 {
	var points int64

	// 1 point for every alphanumeric character in the retailer name
	alphaNumericRegex := regexp.MustCompile(`[a-zA-Z0-9]`)
	points += int64(len(alphaNumericRegex.FindAllString(receipt.Retailer, -1)))

	// 50 points if the total is a round dollar amount with no cents
	if isRoundDollarAmount(receipt.Total) {
		points += 50
	}

	// 25 points if the total is a multiple of 0.25
	if isMultipleOfQuarter(receipt.Total) {
		points += 25
	}

	// 5 points for every two items on the receipt
	points += int64(len(receipt.Items) / 2 * 5)

	// If the trimmed length of the item description is a multiple of 3,
	// multiply the price by 0.2 and round up to the nearest integer
	for _, item := range receipt.Items {
		if len(strings.TrimSpace(item.ShortDescription))%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 64)
			points += int64(math.Ceil(price * 0.2))
		}
	}

	// 6 points if the day in the purchase date is odd
	if isOddDay(receipt.PurchaseDate) {
		points += 6
	}

	// 10 points if the time of purchase is after 2:00pm and before 4:00pm
	if isBetweenTwoAndFour(receipt.PurchaseTime) {
		points += 10
	}

	return points
}

func isRoundDollarAmount(total string) bool {
	amount, _ := strconv.ParseFloat(total, 64)
	return amount == math.Floor(amount)
}

func isMultipleOfQuarter(total string) bool {
	amount, _ := strconv.ParseFloat(total, 64)
	return math.Mod(amount*100, 25) == 0
}

func isOddDay(date string) bool {
	t, _ := time.Parse("2006-01-02", date)
	return t.Day()%2 != 0
}

func isBetweenTwoAndFour(timeStr string) bool {
	t, _ := time.Parse("15:04", timeStr)
	return t.Hour() >= 14 && t.Hour() < 16
}

func validateReceipt(r Receipt) error {
	// Validate retailer (non-empty, matches pattern).
	if matched, _ := regexp.MatchString(`^(\s*\S+\s*)+$`, r.Retailer); !matched {
		return errors.New("invalid retailer")
	}

	// Validate purchaseDate (format: date)
	if _, err := time.Parse("2006-01-02", r.PurchaseDate); err != nil {
		return errors.New("invalid purchase date")
	}

	// Validate purchaseTime (format: time)
	if _, err := time.Parse("15:04", r.PurchaseTime); err != nil {
		return errors.New("invalid purchase time")
	}

	// Validate total (pattern: ^\d+\.\d{2}$)
	if matched, _ := regexp.MatchString(`^\d+\.\d{2}$`, r.Total); !matched {
		return errors.New("invalid total format")
	}

	// Validate items
	if len(r.Items) == 0 {
		return errors.New("no items provided")
	}
	for _, item := range r.Items {
		if err := validateItem(item); err != nil {
			return err
		}
	}

	return nil
}

func validateItem(i Item) error {
	// Validate shortDescription (pattern: ^[\w\s\-]+$)
	if matched, _ := regexp.MatchString(`^[\w\s\-]+$`, i.ShortDescription); !matched {
		return errors.New("invalid item short description")
	}

	// Validate price (pattern: ^\d+\.\d{2}$)
	if matched, _ := regexp.MatchString(`^\d+\.\d{2}$`, i.Price); !matched {
		return errors.New("invalid item price format")
	}

	return nil
}
