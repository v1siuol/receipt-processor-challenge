package main

import (
	"errors"
	"regexp"
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
	id := generateID()
	s.lock.Lock()
	defer s.lock.Unlock()
	s.receiptPoints[id] = calcPoints(receipt)
	return id, nil
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
	// TODO
	var points int64
	// for _, item := range receipt.Items {
	// 	points += int64(item.Price)
	// }
	return points
}

func validateReceipt(r Receipt) error {
	// Validate retailer (non-empty, matches pattern)
	if matched, _ := regexp.MatchString(`^\S+$`, r.Retailer); !matched {
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
