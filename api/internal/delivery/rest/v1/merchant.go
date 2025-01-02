package v1

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"infra/api/internal/domain"
	"infra/api/internal/infra/postgres"
	"infra/api/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (h *Handler) merchantInit(c *gin.Context) {
	var data struct {
		// MerchantId   string `json:"merchant_id" validate:"required"`
		MerchantName string `json:"merchant_name" validate:"required,min=1,max=32,alphanum" `
		// ApiKey       string `json:"api_key" validate:"required"`
		// Mnemonic string `json:"mnemonic" validate:"required"`
	}
	merchantId := uuid.NewString()

	errid := logger.GenErrorId()

	if err := c.ShouldBindJSON(&data); err != nil {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, errid)
		h.log.Debug("bind json error: " + err.Error())
		return
	}

	v := validator.New()

	if err := v.Struct(data); err != nil {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, errid)
		return
	}

	shaBytes := sha256.Sum256([]byte(data.MerchantName + merchantId))

	apiKey := hex.EncodeToString(shaBytes[:])

	// Check by id
	_, err := h.services.Merchants.FindByID(h.db, merchantId)
	if !postgres.IsNotFound(err) {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgMerchantIdExists, "")
		return
	}

	if err == nil {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgMerchantIdExists, "")
		return
	}

	// Check by name
	_, err = h.services.Merchants.FindByName(h.db, data.MerchantName)
	if !postgres.IsNotFound(err) {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgMerchantNameExists, "")
		return
	}

	if err == nil {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgMerchantNameExists, "")
		return
	}

	err = h.db.Transaction(func(tx *gorm.DB) error {
		// create merchant
		merchant := &domain.Merchants{
			MerchantName: data.MerchantName,
			MerchantID:   merchantId,
			ApiKey:       apiKey,
		}

		err := h.services.Merchants.Create(tx, merchant)
		if err != nil {
			return err
		}

		err = h.services.Balances.Init(merchant)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		// TODO: add merchant logstream
		fmt.Println(err)
		responseErr(c, http.StatusInternalServerError, domain.ErrMsgInternalServerError, errid)
		// TODO: send to logstream
		return
	}

	c.AbortWithStatusJSON(http.StatusOK, responseMerchantCreated{
		Error:      false,
		ApiKey:     apiKey,
		MerchantId: merchantId,
	})

}

func (h *Handler) initMerchantRoutes(g *gin.RouterGroup) {
	g.POST("/merchant/create", h.merchantInit)
}
