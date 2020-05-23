package invoice

import (
	"fmt"

	"github.com/imrenagi/go-payment"
)

type LineItemError struct {
	Code int
}

const (
	LineItemErrInvalidQty = iota
)

func (l LineItemError) Error() string {
	switch l.Code {
	case LineItemErrInvalidQty:
		return "Invalid minimum quantity of the items"
	default:
		return "Unrecognized error code"
	}
}

func (l LineItemError) Unwrap() error {
	switch l.Code {
	case LineItemErrInvalidQty:
		return fmt.Errorf("%s: %w", l.Error(), payment.ErrBadRequest)
	default:
		return fmt.Errorf("%s: %w", l.Error(), payment.ErrInternal)
	}
}

// NewLineItem ...
func NewLineItem(
	name, category, merchant string,
	unitPrice float64,
	currency string,
) *LineItem {
	return &LineItem{
		Name:         name,
		Category:     category,
		MerchantName: merchant,
		Currency:     currency,
		UnitPrice:    unitPrice,
		Qty:          1,
	}
}

// LineItem ...
type LineItem struct {
	payment.Model
	InvoiceID    uint64  `json:"-" gorm:"index:line_item_invoice_id_k"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Category     string  `json:"category"`
	MerchantName string  `json:"merchant_name"`
	Currency     string  `json:"currency"`
	UnitPrice    float64 `json:"unit_price"`
	Qty          int     `json:"qty"`
}

func (LineItem) TableName() string {
	return "invoice_line_items"
}

func (i *LineItem) IncreaseQty() error {
	i.Qty = i.Qty + 1
	return nil
}

func (i *LineItem) DecreaseQty() error {
	if i.Qty < 1 {
		return LineItemError{LineItemErrInvalidQty}
	}
	i.Qty = i.Qty - 1
	return nil
}

func (i LineItem) SubTotal() float64 {
	return i.UnitPrice * float64(i.Qty)
}
