package checkouturl

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/project-flogo/core/activity"
	"github.com/project-flogo/core/data/metadata"
)

func init() {
	_ = activity.Register(&Activity{}, New)
}

var activityMd = activity.ToMetadata(&Settings{}, &Input{}, &Output{})

//New optional factory method, should be used if one activity instance per configuration is desired
func New(ctx activity.InitContext) (activity.Activity, error) {

	s := &Settings{}
	err := metadata.MapToStruct(ctx.Settings(), s, true)
	if err != nil {
		return nil, err
	}

	ctx.Logger().Infof("--> Setting: %s", s)

	act := &Activity{
		Settings: *s,
	}

	return act, nil
}

// Activity is an sample Activity that can be used as a base to create a custom activity
type Activity struct {
	Settings Settings
}

// Metadata returns the activity's metadata
func (a *Activity) Metadata() *activity.Metadata {
	return activityMd
}

// Eval implements api.Activity.Eval - Logs the Message
func (a *Activity) Eval(ctx activity.Context) (done bool, err error) {

	input := &Input{}
	err = ctx.GetInputObject(input)
	if err != nil {
		return true, err
	}

	ctx.Logger().Debugf("Input: %s", input)

	params := make(map[string]string)

	data := input.Payload

	now := time.Now().Unix()
	txId := fmt.Sprintf("go-%d", now)
	orderId := txId

	amount := input.Amount * 100

	params["Title"] = data["title"]

	params["vpc_Merchant"] = a.Settings.Merchant
	params["vpc_AccessCode"] = a.Settings.AccessCode
	params["vpc_Command"] = "pay"
	params["vpc_Version"] = "2"

	params["vpc_Amount"] = fmt.Sprintf("%d", amount)

	params["vpc_Currency"] = data["currency"]
	params["vpc_Locale"] = data["locale"]

	params["vpc_MerchTxnRef"] = txId
	params["vpc_OrderInfo"] = orderId

	params["vpc_ReturnURL"] = data["returnUrl"]

	params["AVS_City"] = data["billingCity"]
	params["AVS_Country"] = data["billingCountry"]
	params["AVS_PostCode"] = data["billingPostCode"]
	params["AVS_StateProv"] = data["billingStateProvince"]
	params["AVS_Street01"] = data["billingStreet"]
	params["AgainLink"] = data["againLink"]

	params["vpc_Customer_Email"] = data["customerEmail"]
	params["vpc_Customer_Id"] = data["customerId"]
	params["vpc_Customer_Phone"] = data["customerPhone"]

	params["vpc_SHIP_City"] = data["deliveryCity"]
	params["vpc_SHIP_Country"] = data["deliveryCountry"]
	params["vpc_SHIP_Provice"] = data["deliveryProvince"] // NOTE: vpc_SHIP_Provice is exact in the sepcs documen
	params["vpc_SHIP_Street01"] = data["deliveryAddress"]
	params["vpc_TicketNo"] = data["clientIp"]

	ctx.Logger().Infof("Payment payload: %s", params)

	u, _ := url.Parse(a.Settings.PaymentGwUrl)
	q := u.Query()

	var secureCode []string

	for k, v := range params {
		if len(v) > 0 {
			q.Set(k, v)
		}
	}

	keys := make([]string, 0, len(params))
	for k, v := range params {
		if len(v) > 0 && (strings.HasPrefix(k, "vpc_") || strings.HasPrefix(k, "user_")) {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	format := "%s=%s"
	for _, k := range keys {
		secureCode = append(secureCode, fmt.Sprintf(format, k, params[k]))
	}

	redirectUrl := fmt.Sprintf("%s?%s", a.Settings.PaymentGwUrl, q.Encode())

	if len(secureCode) > 0 {
		data := strings.Join(secureCode, "&")
		fmt.Println("Secure Code ==> ", data)
		secretHash, _ := Sign(a.Settings.SecretHash, data)

		redirectUrl += "&vpc_SecureHash=" + strings.ToUpper(secretHash)
	}

	if len(redirectUrl) > 0 {
		/* act on str */
		output := &Output{RedirectUrl: redirectUrl}
		err = ctx.SetOutputObject(output)
		if err != nil {
			return true, err
		}

		return true, nil
	} else {
		return true, errors.New("RedirectUrl is empty")
	}

}

func Sign(secret string, data string) (string, error) {
	signedKey, _ := hex.DecodeString(secret)

	hmac := hmac.New(sha256.New, signedKey)
	_, err := hmac.Write([]byte(data))
	if err != nil {
		return "", err
	}
	signature := hex.EncodeToString(hmac.Sum(nil))
	return signature, nil
}
