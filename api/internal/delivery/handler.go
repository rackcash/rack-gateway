package delivery

import (
	"infra/api/internal/config"
	v1 "infra/api/internal/delivery/rest/v1"
	"infra/api/internal/infra/nats"
	"infra/api/internal/logger"
	"infra/api/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	Services  *service.Services
	Db        *gorm.DB
	Config    *config.Config
	Natsinfra *nats.NatsInfra
	Log       logger.Logger
}

func (h *Handler) InitAPI(r *gin.Engine) {
	v1Group := r.Group("/v1")

	v1Handler := v1.NewHandler(h.Services, h.Db, h.Config, h.Natsinfra, h.Log)

	{
		v1Handler.InitRoutes(v1Group)
	}
}

func InitHandler(services *service.Services, db *gorm.DB, config *config.Config, natsinfra *nats.NatsInfra, log logger.Logger) *Handler {
	return &Handler{
		Config:    config,
		Natsinfra: natsinfra,
		Log:       log,
		Services:  services,
		Db:        db,
	}
}
