package xendit

import (
	"github.com/imrenagi/go-payment"
	"github.com/imrenagi/go-payment/invoice"

	"fmt"

	"github.com/xendit/xendit-go/ewallet"
	xinvoice "github.com/xendit/xendit-go/invoice"
)

func NewEwalletRequestFromInvoice(inv *invoice.Invoice) (*ewallet.CreatePaymentParams, error) {

	var reqBuilder ewalletRequestBuilder
	var err error

	req := NewEWalletRequest(inv)

	switch inv.Payment.PaymentType {
	case payment.SourceOvo:
		reqBuilder, err = NewOVO(req)
	case payment.SourceDana:
		reqBuilder, err = NewDana(req)
	case payment.SourceLinkAja:
		reqBuilder, err = NewLinkAja(req)
	default:
		return nil, fmt.Errorf("payment type is not known")
	}
	if err != nil {
		return nil, err
	}

	return reqBuilder.Build()
}

func NewInvoiceRequestFromInvoice(inv *invoice.Invoice) (*xinvoice.CreateParams, error) {

	var reqBuilder invoiceRequestBuilder
	var err error

	req := NewInvoiceRequestBuilder(inv)

	switch inv.Payment.PaymentType {
	case payment.SourceOvo:
		reqBuilder, err = NewOVOInvoice(req)
	case payment.SourceDana:
		reqBuilder, err = NewDanaInvoice(req)
	case payment.SourceLinkAja:
		reqBuilder, err = NewLinkAjaInvoice(req)
	case payment.SourceAlfamart:
		reqBuilder, err = NewAlfamartInvoice(req)
	case payment.SourceBCAVA:
		reqBuilder, err = NewBCAVAInvoice(req)
	case payment.SourceBRIVA:
		reqBuilder, err = NewBRIVAInvoice(req)
	case payment.SourceBNIVA:
		reqBuilder, err = NewBNIVAInvoice(req)
	case payment.SourcePermataVA:
		reqBuilder, err = NewPermataVAInvoice(req)
	case payment.SourceMandiriVA:
		reqBuilder, err = NewMandiriVAInvoice(req)
	case payment.SourceCreditCard:
		reqBuilder, err = NewCreditCardInvoice(req)
	default:
		return nil, fmt.Errorf("payment type is not known")
	}
	if err != nil {
		return nil, err
	}

	return reqBuilder.Build()
}
