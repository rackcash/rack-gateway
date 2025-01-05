package v1

import (
	"fmt"
	"infra/api/internal/domain"
	"infra/api/internal/logger"
	"infra/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/shopspring/decimal"
)

func (h *Handler) currencyConvert(c *gin.Context) {
	var data struct {
		Fiat string `json:"fiat" validate:"required,oneof=rub usd eur"`

		// should be lowercase
		Cryptocurrency string  `json:"cryptocurrency" validate:"required,oneof=eth ton sol"`
		Amount         float64 `json:"amount" validate:"required,amount"`
		ApiKey         string  `json:"api_key" validate:"min=64,max=64"`
	}

	var errid = logger.GenErrorId()

	if err := c.ShouldBindJSON(&data); err != nil {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, "")
		fmt.Println("unmarshal error: " + err.Error())
		return
	}

	v := validator.New()

	v.RegisterValidation("amount", validateAmount)

	if err := v.Struct(data); err != nil {
		validationErrs, err := utils.SafeCast[validator.ValidationErrors](err)
		if err != nil || validationErrs == nil {
			responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, "")
			return
		}
		validationErr := validationErrs[0]
		responseErr(c, http.StatusBadRequest, formatValidationErr(data, data.Cryptocurrency, validationErr), "")
		return
	}

	// upper case, cause we accept lower case, but services only accept upper case
	data.Fiat = strings.ToUpper(data.Fiat)

	amountDecimal := decimal.NewFromFloat(data.Amount)

	exists, err := h.services.Merchants.ApiKeyExists(h.db, data.ApiKey)
	if err != nil {
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("api key exists error : "+err.Error(), errid, logger.NA, amountDecimal, data.Cryptocurrency, c.Request.RequestURI, logger.NA, c.ClientIP())
		return
	}

	if !exists {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgApiKeyNotFound, "")
		return
	}

	rates, err := h.services.Rates.Get(data.Fiat)
	if err != nil {
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("rates get error: "+err.Error(), errid, logger.NA, amountDecimal, data.Fiat, c.Request.RequestURI, logger.NA, c.ClientIP())
		return
	}

	converted, rate, err := h.services.Rates.Convert(amountDecimal, domain.StrToCrypto(data.Cryptocurrency), rates)
	if err != nil { // can return only invalid crypto error
		responseErr(c, http.StatusBadRequest, err.Error(), "")
		return
	}

	responseConverterOK := responseConverterOK{
		Error:          false,
		Fiat:           data.Fiat,
		Amount:         amountDecimal,
		Cryptocurrency: data.Cryptocurrency,
		Converted:      converted,
		Rate:           rate,
	}
	c.AbortWithStatusJSON(http.StatusOK, responseConverterOK)
}

func (h *Handler) currencyRates(c *gin.Context) {
	var data struct {
		Fiat   string `json:"fiat" validate:"required,oneof=rub usd eur"`
		ApiKey string `json:"api_key" validate:"min=64,max=64"`
	}

	var errid = logger.GenErrorId()

	if err := c.ShouldBindJSON(&data); err != nil {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, "")
		fmt.Println("unmarshal error: " + err.Error())
		return
	}

	v := validator.New()

	if err := v.Struct(data); err != nil {
		validationErrs, err := utils.SafeCast[validator.ValidationErrors](err)
		if err != nil || validationErrs == nil {
			responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, "")
			return
		}
		validationErr := validationErrs[0]
		responseErr(c, http.StatusBadRequest, formatValidationErr(data, "", validationErr), "")
		return
	}

	// upper case, cause we accept lower case, but services only accept upper case
	data.Fiat = strings.ToUpper(data.Fiat)

	exists, err := h.services.Merchants.ApiKeyExists(h.db, data.ApiKey)
	if err != nil {
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("api key exists error : "+err.Error(), errid, logger.NA, decimal.Zero, logger.NA, c.Request.RequestURI, logger.NA, c.ClientIP())
		return
	}

	if !exists {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgApiKeyNotFound, "")
		return
	}

	rates, err := h.services.Rates.Get(data.Fiat)
	if err != nil {
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		h.log.TemplInvoiceErr("rates get error: "+err.Error(), errid, logger.NA, decimal.Zero, data.Fiat, c.Request.RequestURI, logger.NA, c.ClientIP())
		return
	}

	// TODO: add more cryptocurrencies
	responseRatesOK := responseRatesOK{
		Error: false,
		Fiat:  data.Fiat,
		Rates: responseRates{
			Eth: rates.Eth,
			Ltc: rates.Ltc,
			Sol: rates.Sol,
			Ton: rates.Ton,
		},
	}

	c.AbortWithStatusJSON(http.StatusOK, responseRatesOK)

}

func (h *Handler) initCurrencyRoutes(g *gin.RouterGroup) {
	g.POST("/currency/convert", h.currencyConvert)
	g.POST("/currency/rates", h.currencyRates)
}
