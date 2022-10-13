package types

import "github.com/galayx-future/costpilot/internal/constants/cloud"

type Granularity string

const (
	Monthly Granularity = "MONTHLY"
	Daily   Granularity = "DAILY"
)

type QueryAccountBillRequest struct {
	BillingCycle     string      `position:"Query" name:"BillingCycle"`
	BillingDate      string      `position:"Query" name:"BillingDate"`
	IsGroupByProduct bool        `position:"Query" name:"IsGroupByProduct"`
	Granularity      Granularity `position:"Query" name:"Granularity"`

	//ProductCode      string           `position:"Query" name:"ProductCode"`
	//PageNum          int `position:"Query" name:"PageNum"`
	//OwnerID          requests.Integer `position:"Query" name:"OwnerID"`
	//BillOwnerId      requests.Integer `position:"Query" name:"BillOwnerId"`
	//PageSize         int `position:"Query" name:"PageSize"`
}

type DataInQueryAccountBill struct {
	BillingCycle string                  `json:"BillingCycle" xml:"BillingCycle"`
	AccountID    string                  `json:"AccountID" xml:"AccountID"`
	TotalCount   int                     `json:"TotalCount" xml:"TotalCount"`
	AccountName  string                  `json:"AccountName" xml:"AccountName"`
	Items        ItemsInQueryAccountBill `json:"Items" xml:"Items"`
}
type ItemsInQueryAccountBill struct {
	Item []AccountBillItem `json:"Item" xml:"Item"`
}

type AccountBillItem struct {
	PipCode          PipCode                `json:"PipCode" xml:"PipCode"`
	ProductName      string                 `json:"ProductName" xml:"ProductName"`
	BillingDate      string                 `json:"BillingDate" xml:"BillingDate"`
	SubscriptionType cloud.SubscriptionType `json:"SubscriptionType" xml:"SubscriptionType"`
	Currency         string                 `json:"Currency" xml:"Currency"`
	PretaxAmount     float64                `json:"PretaxAmount" xml:"PretaxAmount"` // 应付金额
}
