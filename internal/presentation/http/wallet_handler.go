package http

import (
	"backend-challenge-go/internal/application"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type WalletHandler struct {
	createWalletUC *application.CreateWalletUseCase
	transactionUC  *application.TransactionUseCase
	reconciliationUC *application.ReconciliationUseCase
}

func NewWalletHandler(createUC *application.CreateWalletUseCase, transUC *application.TransactionUseCase, reconUC *application.ReconciliationUseCase) *WalletHandler {
	return &WalletHandler{
		createWalletUC: createUC,
		transactionUC:  transUC,
		reconciliationUC: reconUC,
	}
}

type createWalletRequest struct {
	PlayerID       string `json:"playerId"`
	InitialBalance struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"initialBalance"`
}

func (h *WalletHandler) CreateWallet(c echo.Context) error {
	var req createWalletRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	input := application.CreateWalletInput{
		PlayerID:       req.PlayerID,
		InitialBalance: req.InitialBalance.Amount,
		Currency:       req.InitialBalance.Currency,
	}

	wallet, err := h.createWalletUC.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]any{
		"id":       wallet.ID(),
		"playerId": wallet.PlayerID(),
		"balance": map[string]string{
			"amount":   formatBalance(wallet.Balance().Amount()),
			"currency": wallet.Balance().Currency(),
		},
		"version": wallet.Version(),
	})
}

func formatBalance(amount int64) string {
	return fmt.Sprintf("%d.%02d", amount/100, amount%100)
}
)
	}

	wallet, err := h.createWalletUC.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"id": wallet.ID()})
}

type processTransactionRequest struct {
	ProviderID            string `json:"providerId"`
	ExternalTransactionID string `json:"externalTransactionId"`
	ReferenceExternalTransactionID string `json:"referenceExternalTransactionId,omitempty"`
	PlayerID              string `json:"playerId"`
	WalletID              string `json:"walletId"`
	Kind                  string `json:"kind"`
	Money                 struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"money"`
}

func (h *WalletHandler) ProcessTransaction(c echo.Context) error {
	var req processTransactionRequest

	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	idempotencyKey := c.Request().Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Idempotency-Key header is required"})
	}

	input := application.ProcessTransactionInput{
		IdempotencyKey:        idempotencyKey,
		ProviderID:            req.ProviderID,
		ExternalTransactionID: req.ExternalTransactionID,
		ReferenceExternalTransactionID: req.ReferenceExternalTransactionID,
		PlayerID:              req.PlayerID,
		WalletID:              req.WalletID,
		Kind:                  req.Kind,
		Amount:                req.Money.Amount,
		Currency:              req.Money.Currency,
	}

	output, err := h.transactionUC.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"transactionId": output.TransactionID,
		"status":        output.Status,
		"balance": map[string]string{
			"amount":   output.Balance,
			"currency": output.Currency,
		},
		"idempotentReplay": output.IdempotentReplay,
	})
})
	}

	input.IdempotencyKey = c.Request().Header.Get("Idempotency-Key")
	if input.IdempotencyKey == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Idempotency-Key header is required"})
	}

	_, err = h.transactionUC.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "transaction processed successfully",
	})
}


func (h *WalletHandler) Reconcile(c echo.Context) error {
	walletId := c.Param("walletId")
	if walletId == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "walletId path parameter is required"})
	}

	output, err := h.reconciliationUC.Execute(c.Request().Context(), walletId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, output)
}
