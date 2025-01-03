package v1

import (
	"infra/api/internal/config"
	"infra/api/internal/infra/nats"
	"infra/api/internal/logger"
	"infra/api/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	services  *service.Services
	db        *gorm.DB
	config    *config.Config
	Natsinfra *nats.NatsInfra
	log       logger.Logger
}

// type

func (h *Handler) InitRoutes(g *gin.RouterGroup) {
	{
		h.initPubInvoiceRoutes(g)
		h.initPrivInvoiceRoutes(g)

		h.initMerchantRoutes(g)
		h.initFinancesRoutes(g)
	}
}

func NewHandler(services *service.Services, db *gorm.DB, config *config.Config, natsinfra *nats.NatsInfra, log logger.Logger) *Handler {
	return &Handler{
		config:    config,
		Natsinfra: natsinfra,
		log:       log,
		services:  services,
		db:        db,
	}
}
