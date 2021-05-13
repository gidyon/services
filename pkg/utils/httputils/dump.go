package httputils

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

const lineText = "============================================================================================================================================"

func DumpRequest(req *http.Request, header string) {
	bs, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Printf("\n%s\n%s\n%s\n\n%s\n%s\n", lineText, header, lineText, "RESPONSE FAILED: "+err.Error(), lineText)
		return
	}
	fmt.Printf("\n%s\n%s\n%s\n\n%s\n%s\n", lineText, header, lineText, string(bs), lineText)
}

func DumpResponse(res *http.Response, header string) {
	bs, err := httputil.DumpResponse(res, true)
	if err != nil {
		fmt.Printf("\n%s\n%s\n%s\n\n%s\n%s\n", lineText, header, lineText, "RESPONSE FAILED: "+err.Error(), lineText)
		return
	}
	fmt.Printf("\n%s\n%s\n%s\n\n%s\n%s\n", lineText, header, lineText, string(bs), lineText)
}
