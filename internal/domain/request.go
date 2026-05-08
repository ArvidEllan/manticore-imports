package domain

import "time"

type Request struct {
	PK                string    `dynamodbav:"pk" json:"-"`
	RequestID         string    `dynamodbav:"requestId" json:"requestId"`
	Reference         string    `dynamodbav:"reference" json:"reference"`
	CustomerName      string    `dynamodbav:"customerName" json:"customerName"`
	Email             string    `dynamodbav:"email" json:"email"`
	Phone             string    `dynamodbav:"phone" json:"phone"`
	CompanyName       string    `dynamodbav:"companyName,omitempty" json:"companyName,omitempty"`
	ProductName       string    `dynamodbav:"productName" json:"productName"`
	ProductCategory   string    `dynamodbav:"productCategory" json:"productCategory"`
	Quantity          int       `dynamodbav:"quantity" json:"quantity"`
	SourceCountry     string    `dynamodbav:"sourceCountry" json:"sourceCountry"`
	PreferredTimeline string    `dynamodbav:"preferredTimeline,omitempty" json:"preferredTimeline,omitempty"`
	ProductURL        string    `dynamodbav:"productUrl,omitempty" json:"productUrl,omitempty"`
	Notes             string    `dynamodbav:"notes,omitempty" json:"notes,omitempty"`
	Status            string    `dynamodbav:"status" json:"status"`
	CreatedAt         time.Time `dynamodbav:"createdAt" json:"createdAt"`
	UpdatedAt         time.Time `dynamodbav:"updatedAt" json:"updatedAt"`
}

type CreateQuoteRequest struct {
	CustomerName      string `json:"customerName"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	CompanyName       string `json:"companyName"`
	ProductName       string `json:"productName"`
	ProductCategory   string `json:"productCategory"`
	Quantity          int    `json:"quantity"`
	SourceCountry     string `json:"sourceCountry"`
	PreferredTimeline string `json:"preferredTimeline"`
	ProductURL        string `json:"productUrl"`
	Notes             string `json:"notes"`
}
