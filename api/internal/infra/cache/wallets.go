package cache

import "infra/api/internal/domain"

func SaveWallet(invoiceId string, wallet *domain.Wallets) {
	WalletsMap.Store(invoiceId, wallet)
}

func FindWallet(invoiceId string) *domain.Wallets {
	v, ok := WalletsMap.Load(invoiceId)
	if !ok {
		return nil
	}
	return v.(*domain.Wallets)
}
