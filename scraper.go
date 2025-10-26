package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func newSessionClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("error creating cookie jar: %w", err)
	}
	return &http.Client{
		Jar:     jar,
		Timeout: 45 * time.Second, // Increased from 20s to handle high load
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		},
	}, nil
}

func checkLicensePlate(licensePlate, vehicleType string) (*SubmitFormResponse, int, error) {
	// Apply rate limiting
	globalRateLimiter.Wait()
	
	var lastErr error
	for attempt := 1; attempt <= maxCaptchaAttempts; attempt++ {
		result, err := performSingleAttempt(licensePlate, vehicleType)
		if err == nil {
			return result, attempt, nil
		}
		if errors.Is(err, errCaptchaMismatch) {
			lastErr = err
			continue
		}
		return nil, attempt, err
	}

	if lastErr != nil {
		return nil, maxCaptchaAttempts, fmt.Errorf("captcha validation failed after %d attempts", maxCaptchaAttempts)
	}

	return nil, maxCaptchaAttempts, fmt.Errorf("failed to check license plate after %d attempts", maxCaptchaAttempts)
}

func performSingleAttempt(licensePlate, vehicleType string) (*SubmitFormResponse, error) {
	client, err := newSessionClient()
	if err != nil {
		return nil, err
	}

	captcha, err := solveCaptcha(client)
	if err != nil {
		return nil, fmt.Errorf("error solving captcha: %w", err)
	}

	data := url.Values{}
	data.Set("BienKS", licensePlate)
	data.Set("Xe", vehicleType)
	data.Set("captcha", captcha)
	data.Set("ipClient", defaultIPClient)
	data.Set("cUrl", formURL)

	req, err := http.NewRequest("POST", submitURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Referer", formURL)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", csgtURL)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	cleanBody := bytes.TrimSpace(responseBody)
	cleanBody = bytes.TrimPrefix(cleanBody, []byte("\xef\xbb\xbf"))

	var submitResponse SubmitFormResponse
	if err := json.Unmarshal(cleanBody, &submitResponse); err != nil {
		responseString := strings.TrimSpace(string(cleanBody))
		if code, convErr := strconv.Atoi(responseString); convErr == nil {
			if code == 404 {
				return nil, errCaptchaMismatch
			}
			return nil, fmt.Errorf("server returned error code: %d", code)
		}
		return nil, fmt.Errorf("error parsing JSON response: %w", err)
	}

	if submitResponse.Href != "" {
		if details, err := fetchResultDetails(client, submitResponse.Href); err == nil {
			submitResponse.Details = details
		} else {
			log.Printf("warning: unable to read result page: %v", err)
		}
	}

	return &submitResponse, nil
}

func fetchResultDetails(client *http.Client, href string) (*ResultDetails, error) {
	if href == "" {
		return nil, nil
	}

	// Retry logic with exponential backoff
	maxRetries := 3
	var lastErr error
	
	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(retry-1)) * time.Second
			time.Sleep(backoff)
			log.Printf("Retrying fetchResultDetails (attempt %d/%d) for: %s", retry+1, maxRetries, href)
		}
		
		req, err := http.NewRequest("GET", href, nil)
		if err != nil {
			return nil, fmt.Errorf("error creating result request: %w", err)
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Referer", formURL)

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("error fetching result page: %w", err)
			continue // Retry on error
		}
		defer resp.Body.Close()

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("error reading result page: %w", err)
			continue // Retry on error
		}

		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("error parsing result page: %w", err)
		}

		details := &ResultDetails{}
		violations := extractViolations(doc)
		if len(violations) > 0 {
			details.Violations = violations
		}
		if len(details.Violations) == 0 {
			text := normalizeMultiline(doc.Find("#bodyPrint123").Text())
			if text == "" {
				text = normalizeMultiline(doc.Find(".xe_texterror").Text())
			}
			details.Message = text
		}

		if details.Message == "" && len(details.Violations) == 0 {
			return nil, nil
		}

		return details, nil
	}
	
	// All retries failed
	return nil, fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
