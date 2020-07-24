package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
)

func main() {
	http.HandleFunc("/", setCookieHandler)
	http.HandleFunc("/read", readCookieHandler)
	http.ListenAndServe(":9090", nil)
}

func someHandler(w http.ResponseWriter, r *http.Request) {
	c := http.Cookie{
		Name:  "theme",
		Value: "dark",
	}
	http.SetCookie(w, &c)
	w.Write([]byte(time.Now().String()))
}

var hashKey = []byte("very-secret")
var s = securecookie.New(hashKey, nil)

func setCookieHandler(w http.ResponseWriter, r *http.Request) {
	encoded, err := s.Encode("cookie-name", "cookie-value")
	if err == nil {
		cookie := &http.Cookie{
			Name:  "cookie-name",
			Value: encoded,
			Path:  "/",
			// HttpOnly: true,
		}
		http.SetCookie(w, cookie)
		fmt.Fprintln(w, encoded)
	}
}

func readCookieHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("cookie-name"); err == nil {
		var value string
		if err = s.Decode("cookie-name", cookie.Value, &value); err == nil {
			fmt.Fprintln(w, value)
		}
	}
}
