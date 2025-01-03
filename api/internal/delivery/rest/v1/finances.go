package v1

import (
	"encoding/json"
	"fmt"
	"infra/api/internal/domain"
	"infra/api/internal/infra/postgres"
	"infra/api/internal/logger"
	"infra/pkg/nats/natsdomain"
	"infra/pkg/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (h *Handler) withdrawal(c *gin.Context) {
	var data struct {
		MerchantId string `json:"merchant_id" validate:"required"`
		// TODO: address validation
		ToAddress string `json:"to_address" validate:"required"`
		// AmountStr      string `json:"-"`
		Cryptocurrency string `json:"cryptocurrency" validate:"required,oneof=eth ton sol"`

		ApiKey string `json:"api_key" validate:"required"`
		Amount string `json:"amount" validate:"required"`

		AmountDecimal decimal.Decimal `json:"-"` // used after validation

		// Amount float64 `json:"amount" validate:"required,amount"`
	}

	errid := logger.GenErrorId()

	if err := c.ShouldBindJSON(&data); err != nil {
		h.log.Debug("should bind:" + err.Error())
		responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, errid)
		return
	}

	v := validator.New()

	v.RegisterValidation("amount", validateAmount)

	if err := v.Struct(data); err != nil {
		validationErrs, err := utils.SafeCast[validator.ValidationErrors](err)
		if err != nil {
			fmt.Println(err)
			responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, "")
			return
		}
		if validationErrs == nil {
			h.log.Debug("validationErrs == nil")
			responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, "")
			return
		}

		validationErr := validationErrs[0]
		responseErr(c, http.StatusBadRequest, formatValidationErr(data, data.Cryptocurrency, validationErr), "")
		return
	}

	ad, err := decimal.NewFromString(data.Amount)
	if err != nil {
		h.log.Debug("decimal.NewFromString(data.Amount) error: " + err.Error())
		responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, "")
		return
	}

	data.AmountDecimal = ad

	merchant, err := h.services.Merchants.FindByID(h.db, data.MerchantId)
	if err != nil {
		if postgres.IsNotFound(err) {
			h.log.Debug("merchant postgres.IsNotFound(err)")
			responseErr(c, http.StatusBadRequest, domain.ErrMsgMerchantNotFound, "")
			return
		}

		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("find merchant error: "+err.Error(), errid, logger.NA, data.AmountDecimal, data.Cryptocurrency, c.Request.RequestURI, data.MerchantId, c.ClientIP())

	}

	if merchant.ApiKey != data.ApiKey {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgApiKeyInvalid, "")
		return
	}

	balance, err := h.services.Balances.Find(h.db, merchant.MerchantID, data.Cryptocurrency)
	if err != nil {
		if postgres.IsNotFound(err) {
			h.log.Debug("balances postgres.IsNotFound(err)")
			responseErr(c, http.StatusBadRequest, domain.ErrMsgGetBalanceError, "")
			return
		}

		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("find balance error: "+err.Error(), errid, logger.NA, data.AmountDecimal, data.Cryptocurrency, c.Request.RequestURI, data.MerchantId, c.ClientIP())
		return
	}

	if balance.Balance.LessThan(data.AmountDecimal) {
		responseErr(c, http.StatusBadRequest, fmt.Sprintf(domain.ErrMsgInsufficientFundsParams, balance.Balance.String()), "")
		return
	}

	withrawalId := uuid.NewString()

	withdrawalData, err := json.Marshal(natsdomain.ReqMerchantWithdrawal{
		WithdrawalTimestamp: time.Now().String(),
		WithdrawalID:        withrawalId,
		Withdrawal: natsdomain.Withdrawal{
			FromAddress: balance.Address,
			MerchantId:  merchant.MerchantID,
			ToAddress:   data.ToAddress,
			Private:     balance.Private,
			Crypto:      data.Cryptocurrency,
			Amount:      data.AmountDecimal,
		},
	})
	if err != nil {
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("marshal error: "+err.Error(), errid, logger.NA, data.AmountDecimal, data.Cryptocurrency, c.Request.RequestURI, data.MerchantId, c.ClientIP())
		return

	}

	err = h.Natsinfra.JsPublish(natsdomain.SubjJsMerchantWithdrawal.String(), withdrawalData)
	if err != nil {
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("js publish error: "+err.Error(), errid, logger.NA, data.AmountDecimal, data.Cryptocurrency, c.Request.RequestURI, data.MerchantId, c.ClientIP())
		return
	}

	// TODO: insert into withdrawal event table
	if err = h.services.Withdrawals.Create(h.db, &domain.Withdrawals{
		WithdrawalID: withrawalId,
		Amount:       data.AmountDecimal,
		From:         balance.Address,
		To:           data.ToAddress,
		Crypto:       data.Cryptocurrency,
		Status:       domain.WITHDRAWAL_PROCESSING,
	}); err != nil {
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("create withdrawal error: "+err.Error(), errid, logger.NA, data.AmountDecimal, data.Cryptocurrency, c.Request.RequestURI, data.MerchantId, c.ClientIP())
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, responseWithdrawalStarted{
		Error:        false,
		WithdrawalID: withrawalId,
		ToAddress:    data.ToAddress,
		Amount:       data.AmountDecimal.String(),
		// TODO: endpoint with withdrawal info
		Status: fmt.Sprintf("%s://%s/v1/finances/withdrawal/info/%s", h.config.Api.Proto, h.config.Api.Ipv4, withrawalId),
	})

}

func (h *Handler) withdrawalInfo(c *gin.Context) {
	var errid = logger.GenErrorId()

	invoiceId := c.Param("withdrawal_id")
	if invoiceId == "" {
		responseErr(c, http.StatusBadRequest, fmt.Sprintf(domain.ErrMsgParamsBadRequest, "withdrawal id is required"), "")
		return
	}

	withdrawal, err := h.services.Withdrawals.Find(h.db, invoiceId)
	if err != nil {
		if postgres.IsNotFound(err) {
			responseErr(c, http.StatusBadRequest, domain.ErrMsgWithdrawalNotFound, "")
			return
		}

		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("find withdrawal error: "+err.Error(), errid, logger.NA, decimal.Zero, logger.NA, c.Request.RequestURI, logger.NA, c.ClientIP())
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, responseWithdrawalInfo{
		Error:     false,
		ToAddress: withdrawal.To,
		Amount:    withdrawal.Amount.String(),
		Status:    withdrawal.Status.ToString(),
		CreatedAt: withdrawal.CreatedAt.Format("2006-01-02 15:04:05"),
	})

}

func (h *Handler) initFinancesRoutes(g *gin.RouterGroup) {
	g.POST("/finances/withdrawal", h.withdrawal)
	g.POST("/finances/withdrawal/info/:withdrawal_id", h.withdrawalInfo)
}
