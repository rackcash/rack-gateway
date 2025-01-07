package app

import (
	"fmt"
	"infra/api/internal/config"
	"infra/api/internal/delivery"
	"infra/api/internal/infra/nats"
	"infra/api/internal/logger"
	"infra/api/internal/service"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
	"gorm.io/gorm"
)

type App struct {
	Config    *config.Config
	Db        *gorm.DB
	NatsInfra *nats.NatsInfra
	Log       logger.Logger
}

func (app *App) Start() {

	defer func() {
		// app.Js()
		// TODO: close logger
		// app.Log.Close()
	}()

	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(cors.Default())

	services := service.HewServices(app.NatsInfra.Ns, app.Db, app.Log, app.Config)

	app.Autostart(services)

	{
		h := delivery.InitHandler(services, app.Db, app.Config, app.NatsInfra, app.Log)

		h.InitAPI(r)
	}

	eChan := make(chan error)
	interrupt := make(chan os.Signal, 1)

	fmt.Println("internal web is starting")

	go func() {
		err := r.Run(app.Config.Api.Ipv4)
		if err != nil {
			eChan <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-eChan:
		app.Log.TemplHTTPError("app fatal error", app.Config.Api.Ipv4, err)
		return
	case <-interrupt:
		return
	}

}

// start autostart services
func (app *App) Autostart(services *service.Services) {

	fmt.Println("Autostart: run find end invoices")
	services.Invoices.RunFindEnd()

	fmt.Println("Autostart: run invoices check")
	services.Invoices.RunAutostartCheck()

	fmt.Println("Autostart: start process events")
	services.OutboxEvents.StartProcessEvents()
	fmt.Println("Autostart: start wait withdrawal")
	services.GetWithdrawal.StartWaitStatus()

	fmt.Println("Autostart: start wait merchant withdrawal")
	services.GetMerchantWithdrawal.StartWaitStatus()

}
