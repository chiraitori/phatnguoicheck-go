package main

import "errors"

const (
	ocrApiURL          = "https://api.ocr.space/parse/image"
	csgtURL            = "https://www.csgt.vn/"
	captchaURL         = csgtURL + "lib/captcha/captcha.class.php"
	submitURL          = csgtURL + "?mod=contact&task=tracuu_post&ajax"
	formURL            = csgtURL + "tra-cuu-phuong-tien-vi-pham.html"
	maxCaptchaAttempts = 9
)

var (
	apiKey             string // Loaded from .env
	defaultIPClient    = "9.9.9.91"
	userAgent          = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/127.0.0.0 Safari/537.36"
	errCaptchaMismatch = errors.New("captcha mismatch")
)
