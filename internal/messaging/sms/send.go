package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gidyon/micro/v2/utils/errs"
	"github.com/gidyon/services/pkg/api/messaging/sms"
	"google.golang.org/grpc/codes"
)

func sendSmsAT(ctx context.Context, opt *Options, sendRequest *sms.SendSMSRequest) {}

func firstVal(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func (smsAPI *smsAPIServer) sendSmsOnfon(ctx context.Context, sendRequest *sms.SendSMSRequest) {
	url := firstVal(sendRequest.GetAuth().GetApiUrl(), "https://api.onfonmedia.co.ke/v1/sms/SendBulkSMS")
	method := "POST"

	errChan := make(chan error, len(sendRequest.GetSms().GetDestinationPhones()))

	for _, phone := range sendRequest.GetSms().GetDestinationPhones() {
		go func(phone string) {
			payload := strings.NewReader(
				fmt.Sprintf(
					"{\"SenderId\": \"%s\",\"IsUnicode\": true,\"IsFlash\": true,\"MessageParameters\": [{\"Number\": \"%s\",\"Text\": \"%s\"}],\"ApiKey\": \"%s\",\"ClientId\": \"%s\"}",
					firstVal(sendRequest.GetAuth().GetSenderId(), "22031"),
					phone,
					sendRequest.GetSms().GetMessage(),
					firstVal(sendRequest.GetAuth().GetApiKey(), "TS1vLuqaV75unatsBeOLn33oORh7+jbOtwPMDlSkK/k="),
					firstVal(sendRequest.GetAuth().GetClientId(), "dea5c505-e95d-48b2-9e67-c7dcdc4d3aa7"),
				),
			)

			req, err := http.NewRequest(method, url, payload)

			if err != nil {
				errChan <- errs.WrapErrorWithCode(codes.Unavailable, err)
				return
			}

			var cookieString string
			for _, cookie := range sendRequest.GetAuth().GetCookies() {
				cookieString += fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value)
			}
			if cookieString == "" {
				cookieString = "AWSALBTG=JttvGPmDtJ0Fw8eZ7nkjZYiOrC62sR+phfsYr/FUl2OHtAjq8XYaFyoh/MblO0MhzrVzzw5KWJfy0p3BhN9RCb7u8xqFo/lA06YaO+GssiR65HzQyaNbomZyr707xNH1N3cU+lubC3+z5/6IqcWI/YeZPoLxf02UHL42aHeC1Az2lAHicG4=; AWSALBTGCORS=JttvGPmDtJ0Fw8eZ7nkjZYiOrC62sR+phfsYr/FUl2OHtAjq8XYaFyoh/MblO0MhzrVzzw5KWJfy0p3BhN9RCb7u8xqFo/lA06YaO+GssiR65HzQyaNbomZyr707xNH1N3cU+lubC3+z5/6IqcWI/YeZPoLxf02UHL42aHeC1Az2lAHicG4="
			}

			req.Header.Add("Cookie", cookieString)
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("AccessKey", firstVal(sendRequest.GetAuth().GetAccessKey(), "DRzrgIdbmQ3hbD19VwJrwlIWbqmhfxI6"))
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", firstVal(sendRequest.GetAuth().GetAuthToken(), "7SbwNHZNhGUUg1FfIxKNr7oKzbUgNGSQ")))

			res, err := smsAPI.HTTPClient.Do(req)
			if err != nil {
				errChan <- errs.WrapErrorWithCode(codes.Unavailable, err)
				return
			}
			defer res.Body.Close()

			resMap := map[string]interface{}{}
			err = json.NewDecoder(res.Body).Decode(&resMap)
			if err != nil {
				errChan <- errs.FromJSONMarshal(err, "sms response")
				return
			}

			if val, ok := resMap["ErrorCode"]; !ok || (fmt.Sprint(val) != "0") {
				errChan <- errs.WrapMessage(codes.Unavailable, "failed to send sms")
				return
			}

			errChan <- nil
		}(phone)
	}

	for range sendRequest.GetSms().GetDestinationPhones() {
		select {
		case <-ctx.Done():
			return
		case err := <-errChan:
			if err != nil {
				smsAPI.Logger.Errorf("failed to send sms: %v", err)
			}
		}
	}
}
