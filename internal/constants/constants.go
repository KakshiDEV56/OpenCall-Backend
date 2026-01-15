package constants

import "time"

var MaxTime = 24 * time.Hour

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type envKeys struct {
	Env                   string
	ServerAddress         string
	CorsAllowedOrigins    string
	DBDriver              string
	DBHost                string
	DBPort                string
	DBUser                string
	DBPassword            string
	DBName                string
	JWTSecret             string
	RazorpayKeyID         string
	RazorpayKeySecret     string
	RazorpayWebhookSecret string
}

type header struct {
	Origin        string
	ContentLength string
	ContentType   string
	Authorization string
}

var EnvKeys = envKeys{
	Env:                   "ENV",
	ServerAddress:         "SERVER_ADDRESS",
	CorsAllowedOrigins:    "CORS_ALLOWED_ORIGINS",
	DBDriver:              "DB_DRIVER",
	DBHost:                "DB_HOST",
	DBPort:                "DB_PORT",
	DBUser:                "DB_USER",
	DBPassword:            "DB_PASSWORD",
	DBName:                "DB_NAME",
	JWTSecret:             "JWT_SECRET",
	RazorpayKeyID:         "RAZORPAY_KEY_ID",
	RazorpayKeySecret:     "RAZORPAY_KEY_SECRET",
	RazorpayWebhookSecret: "RAZORPAY_WEBHOOK_SECRET",
}

var Headers = header{
	Origin:        "Origin",
	ContentLength: "Content-Length",
	ContentType:   "Content-Type",
	Authorization: "Authorization",
}
