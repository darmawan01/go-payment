package service

import (
	"context"
	"errors"
	"time"

	"github.com/imrenagi/go-payment"
	"github.com/imrenagi/go-payment/gateway/midtrans"
	"github.com/rs/zerolog"
	midgo "github.com/veritrans/go-midtrans"
)

type StoreMidtransCreditCardTokenCommand struct {
	UserID string
	Tokens []MidtransCreditCardToken
}

type MidtransCreditCardToken struct {
	StatusCode string `json:"status_code,omitempty"`
	CardHash   string `json:"cardhash"`
	ID         string `json:"token_id"`
}

type UpdateMidtransTransactionStatusCommand struct {
	Response midgo.Response
}

// HandleMidtransNotification takes care of notification sent by midtrans. This checks the validity of the sign key and the similarity
// between the notification and transaction satus.
func (p *Service) HandleMidtransNotification(ctx context.Context, command UpdateMidtransTransactionStatusCommand) error {

	log := zerolog.Ctx(ctx).
		With().
		Str("function", "PaymentService.HandleMidtransNotification()").
		Str("cmd_order_id", command.Response.OrderID).
		Str("cmd_transaction_id", command.Response.TransactionID).
		Str("cmd_gross_amount", command.Response.GrossAmount).
		Str("cmd_transaction_status", command.Response.TransactionStatus).
		Logger()

	storedStatus, err := p.midTransactionRepository.FindByOrderID(ctx, command.Response.OrderID)
	if err != nil && !errors.Is(err, payment.ErrNotFound) {
		return err
	}

	ttt, err := time.Parse("2006-01-02 15:04:05", command.Response.TransactionTime)
	if err != nil {
		log.Error().Err(err).Msg("cant parse transaction time")
		return payment.ErrInternal
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")
	transactionTime := ttt.In(loc)

	if storedStatus == nil {
		storedStatus = &midtrans.TransactionStatus{
			StatusCode:        command.Response.StatusCode,
			StatusMessage:     command.Response.StatusMessage,
			SignKey:           command.Response.SignKey,
			Bank:              command.Response.Bank,
			FraudStatus:       command.Response.FraudStatus,
			PaymentType:       command.Response.PaymentType,
			OrderID:           command.Response.OrderID,
			TransactionID:     command.Response.TransactionID,
			TransactionTime:   transactionTime,
			TransactionStatus: command.Response.TransactionStatus,
			GrossAmount:       command.Response.GrossAmount,
			MaskedCard:        command.Response.MaskedCard,
			Currency:          command.Response.Currency,
			CardType:          command.Response.CardType,
		}

	} else {
		storedStatus.StatusCode = command.Response.StatusCode
		storedStatus.StatusMessage = command.Response.StatusMessage
		storedStatus.GrossAmount = command.Response.GrossAmount
		storedStatus.FraudStatus = command.Response.FraudStatus
		storedStatus.SignKey = command.Response.SignKey
		storedStatus.TransactionTime = transactionTime
		storedStatus.TransactionStatus = command.Response.TransactionStatus
		storedStatus.TransactionID = command.Response.TransactionID
		storedStatus.PaymentType = command.Response.PaymentType
		storedStatus.MaskedCard = command.Response.MaskedCard
		storedStatus.CardType = command.Response.CardType
		storedStatus.Bank = command.Response.Bank
	}

	if err := storedStatus.HasValidSignKey(p.MidtransGateway.GetServerKey()); err != nil {
		return err
	}

	err = p.midTransactionRepository.Save(ctx, storedStatus)
	if err != nil {
		return err
	}

	err = p.processNotification(ctx, *storedStatus)
	if err != nil {
		log.Error().Err(err).Msg("something wrong when publishing")
		return err
	}

	return nil
}

func (p *Service) processNotification(ctx context.Context, status midtrans.TransactionStatus) error {

	log := zerolog.Ctx(ctx).With().
		Str("transaction_status", status.TransactionStatus).
		Str("payment_type", status.PaymentType).
		Str("fraud_status", status.FraudStatus).
		Logger()

	switch status.TransactionStatus {
	case "capture":
		if status.PaymentType == "credit_card" && status.FraudStatus == "accept" {

			_, err := p.PayInvoice(ctx, status.OrderID, PayInvoiceCommand{
				TransactionID: status.TransactionID,
			})
			if err != nil {
				return err
			}
		} else {
			log.Warn().Msg("transaction captured, potentially fraud")
			return nil
		}
	case "settlement":
		_, err := p.PayInvoice(ctx, status.OrderID, PayInvoiceCommand{
			TransactionID: status.TransactionID,
		})
		if err != nil {
			return err
		}
	case "deny", "expire", "cancel":
		_, err := p.FailInvoice(ctx, status.OrderID)
		if err != nil {
			return err
		}
	case "pending":
		_, err := p.ProcessInvoice(ctx, status.OrderID)
		if err != nil {
			return err
		}
	default:
		log.Warn().Msg("payment status type is unidentified")
		return nil
	}

	return nil
}

// StoreMidtransCreditCardToken stores all tokens to the given userID. userID is UUID generated by midtrans.
func (p *Service) StoreMidtransCreditCardToken(ctx context.Context, command StoreMidtransCreditCardTokenCommand) error {

	if len(command.Tokens) == 0 {
		return nil
	}

	for _, token := range command.Tokens {
		err := p.midCardTokenRepository.Save(ctx, &midtrans.CardToken{
			UserID:   command.UserID,
			CardHash: token.CardHash,
			TokenID:  token.ID,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// GetMidtransCreditCardToken returns all tokens stored which belong to the userID. userID is UUID generated by midtrans client app.
func (p *Service) GetMidtransCreditCardToken(ctx context.Context, userID string) ([]MidtransCreditCardToken, error) {

	ccTokens, err := p.midCardTokenRepository.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	tokens := make([]MidtransCreditCardToken, 0)
	for _, t := range ccTokens {
		tokens = append(tokens, MidtransCreditCardToken{
			CardHash: t.CardHash,
			ID:       t.TokenID,
		})
	}
	return tokens, nil
}
