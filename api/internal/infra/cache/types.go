package cache

import "sync"

type Cache struct {
	Storage sync.Map
}

// Maps for cache
var (
	QrCodeMap  sync.Map
	InvoiceMap sync.Map
	// RatesMap   sync.Map
	WalletsMap sync.Map

	InvoiceCheckMap sync.Map
)

// cache
var (
	RatesCache             = InitStorage()
	InvoiceRateLimitsCache = InitStorage()
	ConfigsCache           = InitStorage()
)
