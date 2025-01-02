package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"infra/api/internal/infra/cache"
	"infra/pkg/utils"

	"github.com/yeqown/go-qrcode/v2"

	"github.com/yeqown/go-qrcode/writer/standard"
)

type QrCodesService struct {
	cache *cache.Cache
}

func NewQrCodesService() *QrCodesService {
	return &QrCodesService{cache: cache.InitStorage()}
}

func (s *QrCodesService) New(content string) (string, error) {
	qr, err := generateQrCode(content)
	if err != nil {
		return "", err
	}

	s.cache.SetNoExp(content, qr)

	return qr, nil

}

func (s *QrCodesService) FindOrNew(content string) (string, error) {
	qr, err := utils.SafeCast[string](s.cache.Load(content))
	if err != nil { // not found
		return s.New(content)
	}
	return qr, nil
}

type smallerCircle struct {
	smallerPercent float64
}

// https://github.com/yeqown/go-qrcode/blob/main/example/with-custom-shape/main.go
func (sc *smallerCircle) DrawFinder(ctx *standard.DrawContext) {
	backup := sc.smallerPercent
	sc.smallerPercent = 1.0
	sc.Draw(ctx)
	sc.smallerPercent = backup
}

func newShape(radiusPercent float64) standard.IShape {
	return &smallerCircle{smallerPercent: radiusPercent}
}

func (sc *smallerCircle) Draw(ctx *standard.DrawContext) {
	w, h := ctx.Edge()
	x, y := ctx.UpperLeft()
	color := ctx.Color()

	// choose a proper radius values
	radius := w / 2
	r2 := h / 2
	if r2 <= radius {
		radius = r2
	}

	// 80 percent smaller
	radius = int(float64(radius) * sc.smallerPercent)

	cx, cy := x+float64(w)/2.0, y+float64(h)/2.0 // get center point
	ctx.DrawCircle(cx, cy, float64(radius))
	ctx.SetColor(color)
	ctx.Fill()
}

type bufferAdaptor struct {
	*bytes.Buffer
}

func (b bufferAdaptor) Close() error {
	return nil
}

func (b bufferAdaptor) Write(p []byte) (int, error) {
	return b.Buffer.Write(p)
}

// returns qr code in base64
func generateQrCode(content string) (string, error) {
	shape := newShape(0.7)
	qrc, err := qrcode.New(content)
	if err != nil {
		fmt.Printf("qrcode.New: %v\n", err)
		return "", err
	}

	b := bufferAdaptor{Buffer: bytes.NewBuffer(nil)}
	w2 := standard.NewWithWriter(b, standard.WithCustomShape(shape))

	if err = qrc.Save(w2); err != nil {
		return "", err
	}

	var qrBytes = make([]byte, b.Len())
	_, err = b.Read(qrBytes)
	if err != nil {
		return "", err
	}

	// os.WriteFile("qr.txt", []byte(base64.RawStdEncoding.EncodeToString(qrBytes)), os.ModePerm)

	return base64.RawStdEncoding.EncodeToString(qrBytes), nil
}
