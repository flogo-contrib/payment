package checkouturl

import (
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/kelseyhightower/envconfig"
	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/support/test"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {

	ref := activity.GetRef(&Activity{})
	act := activity.Get(ref)

	assert.NotNil(t, act)
}

func TestEval(t *testing.T) {

	var rawPayload = `
	{
		"amount":               "900000",
		"clientIp":             "127.0.0.1",
		"locale":               "vn",
		"currency":             "VND",
		"customerId":           "dev@naustud.io",
		"returnUrl":            "http://localhost:11800/payment/onepaydom/callback",
		"title":                "VPC 3-Party"
	}
	`

	var payload map[string]string
	err := json.Unmarshal([]byte(rawPayload), &payload)
	if err != nil {
		log.Print(err)
	}
	fmt.Println("payload: ", payload)

	var s Settings

	err = envconfig.Process("onepay", &s)
	if err != nil {
		log.Fatal(err)
		return
	}

	format := "- Gateway: %s\n- AccessCode: %s\n- Merchant: %s\n"
	fmt.Printf(format, s.PaymentGwUrl, s.AccessCode, s.Merchant)

	act := &Activity{Settings: s}
	tc := test.NewActivityContext(act.Metadata())
	input := &Input{Payload: payload, Amount: 10000}
	err = tc.SetInputObject(input)
	assert.Nil(t, err)

	done, err := act.Eval(tc)
	assert.True(t, done)
	assert.Nil(t, err)

	output := &Output{}
	err = tc.GetOutputObject(output)
	assert.Nil(t, err)
	assert.NotNil(t, output.RedirectUrl)

	fmt.Printf("==> Redirect Url: %s", output.RedirectUrl)
	fmt.Println()
}
