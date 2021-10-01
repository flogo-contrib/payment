package checkouturl

import "github.com/project-flogo/core/data/coerce"

// https://github.com/kelseyhightower/envconfig
type Settings struct {
	PaymentGwUrl string `md:"paymentGwUrl" required:"true" split_words:"true" example:"https://mtf.onepay.vn/onecomm-pay/vpc.op"`
	Merchant     string `md:"merchant,required" required:"true"`
	AccessCode   string `md:"accessCode,required" required:"true" split_words:"true"`
	SecretHash   string `md:"secretHash,required" required:"true" split_words:"true"`
}

type Input struct {
	Payload map[string]string `md:"payload,required"`
	Amount  uint32            `md:"amount,required"`
}

func (r *Input) FromMap(values map[string]interface{}) error {
	intVal, _ := coerce.ToInt32(values["amount"])
	r.Amount = uint32(intVal)

	mapVal, _ := coerce.ToParams(values["payload"])
	r.Payload = mapVal
	return nil
}

func (r *Input) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"amount":  r.Amount,
		"payload": r.Payload,
	}
}

type Output struct {
	RedirectUrl string `md:"redirectUrl"`
}

func (o *Output) FromMap(values map[string]interface{}) error {
	strVal, _ := coerce.ToString(values["redirectUrl"])
	o.RedirectUrl = strVal
	return nil
}

func (o *Output) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"redirectUrl": o.RedirectUrl,
	}
}
