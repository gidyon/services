package auth

// CookieName is name of cookie
func CookieName() string {
	return "_x12HIDs"
}

// Scheme returns authentication scheme
func Scheme() string {
	return "Bearer"
}

// Header returns authentication header
func Header() string {
	return "authorization"
}
