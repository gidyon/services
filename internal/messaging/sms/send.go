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
					firstVal(sendRequest.GetAuth().GetSenderId(), smsAPI.SmsAuth.GetSenderId()),
					phone,
					sendRequest.GetSms().GetMessage(),
					firstVal(sendRequest.GetAuth().GetApiKey(), smsAPI.SmsAuth.GetApiKey()),
					firstVal(sendRequest.GetAuth().GetClientId(), smsAPI.SmsAuth.GetClientId()),
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
				for _, cookie := range smsAPI.SmsAuth.GetCookies() {
					cookieString += fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value)
				}
			}

			req.Header.Add("Cookie", cookieString)
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("AccessKey", firstVal(sendRequest.GetAuth().GetAccessKey(), smsAPI.SmsAuth.GetAccessKey()))
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", firstVal(sendRequest.GetAuth().GetAuthToken(), smsAPI.SmsAuth.GetAuthToken())))

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
