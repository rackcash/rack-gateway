package cache

// Check qr code in cache or not
func IsQrCodeInCache(address string) bool {
	_, ok := QrCodeMap.Load(address)
	return ok
}

func SaveQrCode(address string, qrCode string) {
	QrCodeMap.Store(address, qrCode)
}

// returns qr code from cache
//
// if not found, returns an empty string ("")
func FindQrCode(address string) string {
	qrCode, ok := QrCodeMap.Load(address)
	if !ok {
		return ""
	}
	return qrCode.(string)
}
