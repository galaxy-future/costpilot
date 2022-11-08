package data

import "github.com/galaxy-future/costpilot/internal/constants/cloud"

type ItemInProductBilling struct {
	PipCode          string                 `json:"pip_code"`
	ProductName      string                 `json:"product_name"`
	PretaxAmount     float64                `json:"pretax_amount"`     // 应付金额
	SubscriptionType cloud.SubscriptionType `json:"subscription_type"` // PostPaid | PrePaid
	Currency         string                 `json:"currency"`
}
type ProductBilling struct {
	ProductName string                 `json:"product_name"`
	TotalAmount float64                `json:"total_amount"`
	Items       []ItemInProductBilling `json:"Items"`
}
type DailyBilling struct {
	Day             string                    `json:"day"`              // 20220101
	ProductsBilling map[string]ProductBilling `json:"products_billing"` // map['pip_code'] key = ecs
	TotalAmount     float64                   `json:"total_amount"`
}
type MonthlyBilling struct {
	Month           string                    `json:"month"`            // 202201
	ProductsBilling map[string]ProductBilling `json:"products_billing"` // map['product_name']
	TotalAmount     float64                   `json:"total_amount"`
}
type YearlyBilling struct {
	Year        string  `json:"year"` // 2022
	TotalAmount float64 `json:"total_amount"`
}
type AccountBilling struct {
	AccountName  string                     `json:"account_name"`
	YearsBilling map[string][]YearlyBilling `json:"years_billing"`
	Provider     cloud.Provider             `json:"provider"`
}
