package v1

import (
	"fmt"
	"infra/api/internal/infra/cache"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// const (
// 	MAX_LIFETIME = 4320
// )

// var (
// 	MAX_AMOUNT = decimal.NewFromInt(10 << 20)
// )

const DEFAULT_LIMIT = 150
const EXPIRATION_SECONDS = 30

// returns true if rate limit is exceeded
func invoiceRateLimit(apiKey string, limit int) bool {
	fmt.Println("LIMIT: ", limit)
	var expiration = time.Second * time.Duration(EXPIRATION_SECONDS)

	count := cache.InvoiceRateLimitsCache.LoadOrSet(apiKey, 1, expiration)
	if count == nil {
		fmt.Println("count == nil")
		return true
	}

	countInt, ok := count.(int)
	if !ok {
		fmt.Println("!ok")
		return true
	}

	if countInt > limit {
		return true
	}

	cache.InvoiceRateLimitsCache.Set(apiKey, countInt+1, expiration)
	return false
}

func (h *Handler) adminAccessMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if h.config.PrivateKey != c.Request.Header.Get("Access") {
			responseErr(c, http.StatusUnauthorized, "access denied", "")
			c.Abort()
			return
		}
		c.Next()

	}

}
