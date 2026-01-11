package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IfedayoAwe/payment-processing-service/db/gen"
	"github.com/IfedayoAwe/payment-processing-service/models"
	"github.com/IfedayoAwe/payment-processing-service/pkg/money"
	"github.com/IfedayoAwe/payment-processing-service/providers"
	"github.com/IfedayoAwe/payment-processing-service/queue"
	"github.com/IfedayoAwe/payment-processing-service/utils"
)

type PayoutWorker interface {
	ProcessPayoutJob(ctx context.Context, job *queue.Job) error
	StartWorker(ctx context.Context) error
}

type payoutWorker struct {
	queries  *gen.Queries
	provider *providers.Processor
	queue    queue.Queue
}

func (s *Services) PayoutWorker() PayoutWorker {
	return &payoutWorker{
		queries:  s.queries,
		provider: s.provider,
		queue:    s.queue,
	}
}

func (pw *payoutWorker) ProcessPayoutJob(ctx context.Context, job *queue.Job) error {
	var payload queue.PayoutJobPayload
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payout job payload: %w", err)
	}

	transaction, err := pw.queries.GetTransactionByID(ctx, payload.TransactionID)
	if err != nil {
		return fmt.Errorf("get transaction: %w", err)
	}

	if transaction.Status != string(models.TransactionStatusPending) {
		return nil
	}

	currency, err := money.ParseCurrency(payload.Currency)
	if err != nil {
		return fmt.Errorf("invalid currency: %w", err)
	}

	bankAccount, err := pw.getBankAccount(ctx, payload.BankAccountID)
	if err != nil {
		return fmt.Errorf("get bank account: %w", err)
	}

	payoutReq := providers.PayoutRequest{
		Amount: money.NewMoney(payload.Amount, currency),
		Destination: providers.BankAccount{
			BankName:      bankAccount.BankName,
			AccountNumber: bankAccount.AccountNumber,
			AccountName:   getStringOrEmpty(bankAccount.AccountName),
			Currency:      currency,
		},
		Metadata: map[string]string{
			"transaction_id":  payload.TransactionID,
			"bank_account_id": payload.BankAccountID,
		},
	}

	payoutResp, err := pw.provider.SendPayout(ctx, payoutReq)
	if err != nil {
		failErr := pw.failTransaction(ctx, payload.TransactionID, fmt.Sprintf("provider payout failed: %v", err))
		if failErr != nil {
			return fmt.Errorf("provider payout failed: %w (also failed to mark transaction failed: %v)", err, failErr)
		}
		return err
	}

	err = pw.queries.UpdateTransactionWithProvider(ctx, gen.UpdateTransactionWithProviderParams{
		ProviderName:      sql.NullString{String: payoutResp.ProviderName, Valid: true},
		ProviderReference: sql.NullString{String: string(payoutResp.ProviderRef), Valid: true},
		Status:            string(models.TransactionStatusCompleted),
		ID:                payload.TransactionID,
	})
	if err != nil {
		return fmt.Errorf("update transaction: %w", err)
	}

	webhookPayload, err := json.Marshal(map[string]interface{}{
		"transaction_id":     payload.TransactionID,
		"provider_reference": string(payoutResp.ProviderRef),
		"status":             "completed",
		"amount":             payload.Amount,
		"currency":           payload.Currency,
	})
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	webhookJobPayload := queue.WebhookJobPayload{
		ProviderName:      payoutResp.ProviderName,
		EventType:         "payout.completed",
		ProviderReference: string(payoutResp.ProviderRef),
		TransactionID:     &payload.TransactionID,
		Payload:           webhookPayload,
	}

	if err := pw.queue.Enqueue(ctx, queue.JobTypeWebhook, webhookJobPayload); err != nil {
		return fmt.Errorf("enqueue webhook job: %w", err)
	}

	return nil
}

func (pw *payoutWorker) StartWorker(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := pw.queue.Process(ctx, queue.JobTypePayout, pw.ProcessPayoutJob, 5*time.Second); err != nil {
				utils.Logger.Error().Err(err).Str("job_type", string(queue.JobTypePayout)).Msg("error processing payout job")
			}
		}
	}
}

func (pw *payoutWorker) failTransaction(ctx context.Context, transactionID string, reason string) error {
	err := pw.queries.UpdateTransactionFailure(ctx, gen.UpdateTransactionFailureParams{
		Status:        string(models.TransactionStatusFailed),
		FailureReason: sql.NullString{String: reason, Valid: true},
		ID:            transactionID,
	})
	return err
}

func (pw *payoutWorker) getBankAccount(ctx context.Context, bankAccountID string) (*models.BankAccount, error) {
	bankAccount, err := pw.queries.GetBankAccountByID(ctx, bankAccountID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.NotFoundErr("bank account not found")
		}
		return nil, utils.ServerErr(fmt.Errorf("get bank account: %w", err))
	}

	var accountName *string
	if bankAccount.AccountName.Valid {
		accountName = &bankAccount.AccountName.String
	}

	return &models.BankAccount{
		ID:            bankAccount.ID,
		UserID:        bankAccount.UserID,
		BankName:      bankAccount.BankName,
		BankCode:      bankAccount.BankCode,
		AccountNumber: bankAccount.AccountNumber,
		AccountName:   accountName,
		Currency:      bankAccount.Currency,
		Provider:      bankAccount.Provider,
		CreatedAt:     bankAccount.CreatedAt,
		UpdatedAt:     bankAccount.UpdatedAt,
	}, nil
}

func getStringOrEmpty(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
