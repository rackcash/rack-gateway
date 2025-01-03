package v1

import (
	"fmt"
	"infra/api/internal/domain"
	"infra/pkg/utils"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/shopspring/decimal"
)

// lifetime - int - max 4320
// amount - int - max 10000000
// description - string
// api key - string - min 64, max 64
//  webhook - string - https://

type limit struct {
	Min decimal.Decimal
	Max decimal.Decimal
}

// TODO: change
var amountLimits = map[string]limit{
	"eth": {Min: decimal.NewFromFloat(0), Max: decimal.NewFromInt(100000000)},
	"ton": {Min: decimal.NewFromFloat(0), Max: decimal.NewFromInt(100000000)},
	"sol": {Min: decimal.NewFromFloat(0), Max: decimal.NewFromInt(100000000)},
}

type NewInvoiceData struct {
	Lifetime       int     `json:"lifetime" validate:"required,gte=0,lte=4320"`
	Cryptocurrency string  `json:"cryptocurrency" validate:"required,oneof=eth ton sol"`
	AmountFloat    float64 `json:"amount" validate:"required"`
	ApiKey         string  `json:"api_key" validate:"min=64,max=64"` // sha256
	Webhook        string  `json:"webhook" validate:"webhook,max=60"`

	Amount decimal.Decimal `json:"-"` // used after validation
}

// checks the validity of data in query
// returns false if there is an error
func filterQuery(c *gin.Context) (*NewInvoiceData, bool) {

	var data NewInvoiceData
	err := c.ShouldBindJSON(&data)
	if err != nil {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, "")
		return nil, false
	}

	v := validator.New()

	v.RegisterValidation("amount", validateAmount)
	v.RegisterValidation("webhook", validateWebhook)
	err = v.Struct(data)
	if err == nil {
		data.Amount = decimal.NewFromFloat(data.AmountFloat)

		return &data, true
	}

	validationErrs, err := utils.SafeCast[validator.ValidationErrors](err)
	if err != nil {
		fmt.Println(err)
		responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, "")
		return nil, false
	}
	if validationErrs == nil {
		responseErr(c, http.StatusBadRequest, domain.ErrMsgBadRequest, "")
		return nil, false
	}

	validationErr := validationErrs[0]
	responseErr(c, http.StatusBadRequest, formatValidationErr(data, data.Cryptocurrency, validationErr), "")

	return nil, false

}

func validateAmount(fl validator.FieldLevel) bool {

	obj := fl.Parent()
	amount := fl.Field().Float()

	// TODO: fix
	amountCurrency := obj.FieldByName("cryptocurrency")
	if !amountCurrency.IsValid() {
		fmt.Println("Invalid field by name: cryptocurrency")
		return false
	}

	limit, ok := amountLimits[amountCurrency.String()]
	if !ok {
		return false
	}

	amountDecimal := decimal.NewFromFloat(amount)

	if amountDecimal.LessThan(limit.Min) || amountDecimal.GreaterThan(limit.Max) {
		return false
	}

	return true
}

func validateWebhook(fl validator.FieldLevel) bool {
	if fl.Field().String() == "" { // webhook is not set
		return true
	}

	if len(fl.Field().String()) <= 8 {
		return false
	}
	// TODO: uncomment
	// if !strings.HasPrefix(fl.Field().String(), "https://") { // is https
	// 	return false
	// }
	if !strings.Contains(fl.Field().String(), ".") { // has dot
		return false
	}

	_, err := url.ParseRequestURI(fl.Field().String())
	return err == nil
}

func formatValidationErr(data any, cryptocurrency string, err validator.FieldError) string {
	jsonTag := getJSONTag(data, err.Field())

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("field '%s' is required", jsonTag)
	case "oneof":
		return fmt.Sprintf("field '%s' must be one of '%s'", jsonTag, err.Param())
	case "min":
		return fmt.Sprintf("field '%s' must be at least %s characters long", jsonTag, err.Param())
	case "max":
		return fmt.Sprintf("field '%s' must be at most %s characters long", jsonTag, err.Param())
	case "gte":
		return fmt.Sprintf("field '%s' must be greater than or equal to %s", jsonTag, err.Param())
	case "lte":
		return fmt.Sprintf("field '%s' must be less than or equal to %s", jsonTag, err.Param())
	//  custom tags
	case "webhook":
		return fmt.Sprintf("field '%s' must be a valid HTTPS url", jsonTag)
	case "amount":
		limit, ok := amountLimits[cryptocurrency]
		if !ok {
			var currencyList string
			for k := range amountLimits {
				currencyList += k + " "
			}
			return fmt.Sprintf("field cryptocurrency must be one of '%s'", currencyList)
		}
		return fmt.Sprintf("field '%s' must be greater than %s and less than %s", jsonTag, limit.Min, limit.Max)

	default:
		return fmt.Sprintf("invalid field '%s'", jsonTag)
	}

}

func getJSONTag(structType any, fieldName string) string {
	typ := reflect.TypeOf(structType)
	field, _ := typ.FieldByName(fieldName)
	tag := field.Tag.Get("json")
	if tag == "" {
		return fieldName
	}
	return tag
}
