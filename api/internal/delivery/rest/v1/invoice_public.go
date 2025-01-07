// PUBLIC INVOICE ROUTES

package v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"infra/api/internal/domain"
	"infra/api/internal/infra/postgres"
	"infra/api/internal/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// lifetime - int - max 4320
// amount - int
// description - string
// api key - string

// /{version}/invoice/create
func (h *Handler) invoiceCreate(c *gin.Context) {
	var errid = logger.GenErrorId()
	invoiceData, ok := filterQuery(c)
	if !ok || invoiceData == nil {
		return
	}

	endTimestamp := time.Now().Add(time.Duration(invoiceData.Lifetime) * time.Minute).Unix()

	merchant, err := h.services.Merchants.FindByApiKey(h.db, invoiceData.ApiKey)
	if err != nil {
		if postgres.IsNotFound(err) {
			responseErr(c, http.StatusBadRequest, domain.ErrMsgApiKeyNotFound, "")
		} else {
			responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
			h.log.TemplInvoiceErr("find merchant by api key error : "+err.Error(), errid, logger.NA, invoiceData.Amount, invoiceData.Cryptocurrency, c.Request.RequestURI, logger.NA, c.ClientIP())
		}
		return
	}

	// check rate limit

	isRateLimited := invoiceRateLimit(invoiceData.ApiKey, 200) // default limit
	if isRateLimited {
		responseErr(c, http.StatusTooManyRequests, domain.ErrMsgRateLimitExceeded, "")
		return
	}

	invoice_id := uuid.NewString()

	err = h.services.Balances.Init(merchant)
	if err != nil {
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("init balance error: "+err.Error(), errid, invoice_id, invoiceData.Amount, invoiceData.Cryptocurrency, c.Request.RequestURI, merchant.MerchantID, c.ClientIP())
		return
	}

	err = h.services.Invoices.Create(h.db, &domain.Invoices{
		InvoiceID:      invoice_id,
		MerchantID:     merchant.MerchantID,
		EndTimestamp:   endTimestamp,
		Status:         domain.STATUS_NOT_PAID,
		Amount:         invoiceData.Amount,
		Cryptocurrency: invoiceData.Cryptocurrency,
		Webhook:        invoiceData.Webhook,
	})
	if err != nil {
		h.log.TemplInvoiceErr("invoice create error: "+err.Error(), errid, invoice_id, invoiceData.Amount, invoiceData.Cryptocurrency, c.Request.RequestURI, merchant.MerchantID, c.ClientIP())
		return
	}

	// create wallet

	wallet, err := h.services.Wallets.CreateAndSave(invoice_id, merchant.MerchantID, domain.StrToCrypto(invoiceData.Cryptocurrency))
	if err != nil {
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("create wallet error: "+err.Error(), errid, invoice_id, invoiceData.Amount, invoiceData.Cryptocurrency, c.Request.RequestURI, merchant.MerchantID, c.ClientIP())
		return
	}

	invoice, err := h.services.Invoices.FindAndSaveToCache(invoice_id)
	if err != nil {
		status := domain.GetStatusByErr(err)
		responseErr(c, status, err.Error(), errid)
		return
	}

	_, err = h.services.QrCodes.New(wallet.Address)
	if err != nil {
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("qr code new error: "+err.Error(), errid, invoice_id, invoice.Amount, invoice.Cryptocurrency, c.Request.RequestURI, invoice.MerchantID, c.ClientIP())
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Until(time.Unix(invoice.EndTimestamp, 0)))
	go h.services.Invoices.RunCheck(ctx, cancel, invoice, wallet.Address)

	c.AbortWithStatusJSON(http.StatusOK, responseInvoiceCreated{
		Error: false,
		Invoice: responseInvoiceCreatedInfo{
			Id: invoice_id,
			Wallet: responseInvoiceCreatedWallet{
				// TODO: fix
				QrCode:         fmt.Sprintf("%s://%s/v1/invoice/qr-code/%s", h.config.Api.Proto, h.config.Api.Ipv4, invoice_id),
				Address:        wallet.Address,
				AmountToPay:    invoiceData.Amount,
				Cryptocurrency: invoiceData.Cryptocurrency,
				//
			},
		},
	})

	h.log.TemplInvoiceInfo("new invoice created", errid, invoice_id, invoiceData.Amount, invoiceData.Cryptocurrency, c.Request.RequestURI, merchant.MerchantID, c.ClientIP())
}

// POST /invoice/info
func (h *Handler) invoiceInfo(c *gin.Context) {
	// TODO: validation
	var data struct {
		InvoiceId string `json:"invoice_id"`
		ApiKey    string `json:"api_key"`
	}

	var errid = logger.GenErrorId()

	if err := c.ShouldBindJSON(&data); err != nil {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, "")

		// TODO: test
		fmt.Println("unmarshal error: " + err.Error())
		// app.Log.TemplInvoiceLog(logger.LL_ERROR, "unmarshal error: "+err.Error(), helpers.NA, decimal.NewFromInt(0), helpers.NA, c.Request.RequestURI)
		return
	}

	if data.InvoiceId == "" {
		responseErr(c, http.StatusBadRequest, fmt.Sprintf(domain.ErrMsgParamsBadRequest, domain.ErrParamEmptyInvoiceId), "")
		return
	}

	if data.ApiKey == "" {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgApiKeyInvalid, "")
		return
	}

	invoice, err := h.services.Invoices.FindGlobal(h.db, data.InvoiceId)
	if err != nil {
		// responses.ErrWithMsg(c, responses.ErrMsgInternalServerError, http.StatusInternalServerError)
		responseErr(c, domain.GetStatusByErr(err), err.Error(), errid)
		return
	}

	merchant, err := h.services.Merchants.FindByApiKey(h.db, data.ApiKey)
	if err != nil {
		if postgres.IsNotFound(err) {
			responseErr(c, http.StatusBadRequest, domain.ErrMsgApiKeyNotFound, "")
			return
		}

		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("find by api key error: "+err.Error(), errid, data.InvoiceId, decimal.Zero, logger.NA, c.Request.RequestURI, invoice.MerchantID, c.ClientIP())
		return
	}

	if merchant.MerchantID != invoice.MerchantID {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgApiKeyInvalid, "")
		return
	}

	fmt.Println("INVOICE ID", data.InvoiceId)

	var response = domain.ResponseInvoiceInfo{
		Id:             invoice.InvoiceID,
		Amount:         invoice.Amount.String(),
		Cryptocurrency: invoice.Cryptocurrency,
		IsPaid:         invoice.Status.IsPaid(),
		Status:         invoice.Status.ToString(),
		CreatedAt:      invoice.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if time.Now().Unix() > invoice.EndTimestamp && invoice.Status.IsNotPaid() && !invoice.Status.IsCancelled() {
		response.Status = "end"
	}

	responseM, err := json.Marshal(&response)
	if err != nil {
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("/info/ marshal error: "+err.Error(), errid, data.InvoiceId, decimal.Zero, logger.NA, c.Request.RequestURI, invoice.MerchantID, c.ClientIP())
		return
	}

	c.Data(http.StatusOK, "application/json", responseM)

}

func (h *Handler) qrCode(c *gin.Context) {
	var errid = logger.GenErrorId()

	invoiceId := c.Param("invoice_id")
	if invoiceId == "" {
		responseErr(c, http.StatusBadRequest, fmt.Sprintf(domain.ErrMsgParamsBadRequest, "invoice id is required"), "")
		return
	}

	wallet, err := h.services.Wallets.FindByInvoiceID(h.db, invoiceId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			responseErr(c, http.StatusBadRequest, domain.ErrMsgInvalidInvoiceId, "")
			return
		}
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("find invoice by id error: "+err.Error(), errid, invoiceId, decimal.Zero, logger.NA, c.Request.RequestURI, logger.NA, c.ClientIP())
		return
	}

	qrCode, err := h.services.QrCodes.FindOrNew(wallet.Address)
	if err != nil {
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("qr code find or new error: "+err.Error(), errid, invoiceId, decimal.Zero, wallet.Crypto, c.Request.RequestURI, wallet.MerchantID, c.ClientIP())
		return
	}

	c.Data(http.StatusOK, "image/png", []byte(qrCode))
}

func (h *Handler) invoiceCancel(c *gin.Context) {
	var data struct {
		InvoiceId string `json:"invoice_id"`
		ApiKey    string `json:"api_key"`
	}

	var errid = logger.GenErrorId()

	if err := c.ShouldBindJSON(&data); err != nil {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, "")
		fmt.Println("unmarshal error: " + err.Error())
		return
	}

	if data.InvoiceId == "" {
		responseErr(c, http.StatusBadRequest, fmt.Sprintf(domain.ErrMsgParamsBadRequest, domain.ErrParamEmptyInvoiceId), "")
		return
	}

	if data.ApiKey == "" {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgApiKeyInvalid, "")
		return
	}

	exists, err := h.services.Merchants.ApiKeyExists(h.db, data.ApiKey)
	if err != nil {
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("api key exists error: "+err.Error(), errid, data.InvoiceId, decimal.Zero, logger.NA, c.Request.RequestURI, logger.NA, c.ClientIP())
		return
	}

	if !exists {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgApiKeyNotFound, "")
		return
	}

	if err = h.services.Invoices.Cancel(data.InvoiceId); err != nil {
		var _errid = ""
		var errmsg = err.Error()

		code := domain.GetStatusByErr(err)
		if code == http.StatusInternalServerError {
			_errid = errid
			errmsg = domain.ErrMsgInternalServerError
			h.log.TemplInvoiceErr("invoice cancel error: "+err.Error(), errid, data.InvoiceId, decimal.Zero, logger.NA, c.Request.RequestURI, logger.NA, c.ClientIP())
		}
		responseErr(c, code, errmsg, _errid)
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, responseInvoiceCancelled{
		Error: false,
	})

}

func (h *Handler) initPubInvoiceRoutes(g *gin.RouterGroup) {
	g.POST("/invoice/create", h.invoiceCreate)
	g.POST("/invoice/info", h.invoiceInfo)
	g.POST("/invoice/cancel", h.invoiceCancel)

	g.GET("/invoice/qr-code/:invoice_id", h.qrCode)
}
