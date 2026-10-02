package http

import (
	"backend-challenge-go/internal/application"
	"net/http"

	"github.com/labstack/echo/v4"
)

type WalletHandler struct {
	createWalletUC *application.CreateWalletUseCase
	transactionUC  *application.TransactionUseCase
}

func NewWalletHandler(createUC *application.CreateWalletUseCase, transUC *application.TransactionUseCase) *WalletHandler {
	return &WalletHandler{
		createWalletUC: createUC,
		transactionUC:  transUC,
	}
}

func (h *WalletHandler) CreateWallet(c echo.Context) error {
	var input application.CreateWalletInput

	err := c.Bind(&input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	wallet, err := h.createWalletUC.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"id": wallet.ID()})
}

func (h *WalletHandler) ProcessTransaction(c echo.Context) error {
	var input application.ProcessTransactionInput

	err := c.Bind(&input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	_, err = h.transactionUC.Execute(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusUnprocessableEntity, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "transaction processed successfully",
	})
}
