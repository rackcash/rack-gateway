// PRIVATE INVOICE ROUTES

package v1

import (
	"fmt"
	"infra/api/internal/config"

	"github.com/gin-gonic/gin"
)

func (h *Handler) updateProxyList(c *gin.Context) {
	// h.config.PrivateKey

	fmt.Println("UPDATE PROXY LIST")

	h.services.WebhookSender.UpdateList(config.GetProxyList(h.config.ProxyPath))
	c.JSON(200, gin.H{
		"ok": true,
	})

}

func (h *Handler) getProxyList(c *gin.Context) {
	c.JSON(200, gin.H{
		"proxies": h.services.WebhookSender.GetList(),
	})
}

func (h *Handler) initPrivInvoiceRoutes(g *gin.RouterGroup) {
	g.POST("/webhook/updateProxyList", h.adminAccessMiddleware(), h.updateProxyList)
	g.POST("/webhook/getProxyList", h.adminAccessMiddleware(), h.getProxyList)
}
