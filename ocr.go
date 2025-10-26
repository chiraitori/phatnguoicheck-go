package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

func solveCaptcha(client *http.Client) (string, error) {
	resp, err := client.Get(captchaURL)
	if err != nil {
		return "", fmt.Errorf("error downloading captcha: %w", err)
	}
	defer resp.Body.Close()

	imageData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading image data: %w", err)
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("error decoding image: %w", err)
	}

	// Convert to grayscale
	grayscaleImg := imaging.Grayscale(img)

	// Adjust contrast
	contrastImg := imaging.AdjustContrast(grayscaleImg, 20)

	// Try Tesseract first
	text, err := solveWithTesseract(contrastImg)
	if err == nil && text != "" {
		log.Printf("Tesseract OCR succeeded: %s", text)
		return text, nil
	}

	log.Printf("Tesseract failed (%v), trying OCR.space API as fallback...", err)

	// Fallback to OCR.space API
	text, err = solveWithOCRAPI(contrastImg)
	if err != nil {
		return "", fmt.Errorf("both OCR methods failed: %w", err)
	}

	return text, nil
}

func solveWithOCRAPI(img image.Image) (string, error) {
	// Create a temporary file to save image for encoding
	tmpFile, err := ioutil.TempFile("", "captcha-ocr-*.jpg")
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Save image to temp file
	if err := imaging.Save(img, tmpPath, imaging.JPEGQuality(95)); err != nil {
		return "", fmt.Errorf("error saving temp image: %w", err)
	}

	// Read back the JPEG data
	jpegData, err := ioutil.ReadFile(tmpPath)
	if err != nil {
		return "", fmt.Errorf("error reading temp image: %w", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(jpegData)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("apikey", apiKey)
	writer.WriteField("base64image", "data:image/jpeg;base64,"+base64Image)
	writer.Close()

	req, err := http.NewRequest("POST", ocrApiURL, body)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	ocrClient := &http.Client{Timeout: 20 * time.Second}
	res, err := ocrClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	responseBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	var ocrResponse OCRResponse
	if err := json.Unmarshal(responseBody, &ocrResponse); err != nil {
		return "", fmt.Errorf("error parsing JSON: %w", err)
	}

	if ocrResponse.IsErroredOnProcessing {
		return "", fmt.Errorf("OCR processing error: %s", ocrResponse.ErrorMessage)
	}

	if len(ocrResponse.ParsedResults) > 0 {
		return strings.TrimSpace(ocrResponse.ParsedResults[0].ParsedText), nil
	}

	return "", fmt.Errorf("no text found")
}

func solveWithTesseract(img image.Image) (string, error) {
	// Save to temporary file
	tmpFile, err := ioutil.TempFile("", "captcha-*.png")
	if err != nil {
		return "", fmt.Errorf("error creating temp file: %w", err)
	}
	defer tmpFile.Close()
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	// Save processed image
	if err := imaging.Save(img, tmpFile.Name()); err != nil {
		return "", fmt.Errorf("error saving temp image: %w", err)
	}

	// Run Tesseract
	cmd := exec.Command("tesseract", tmpFile.Name(), "stdout",
		"--psm", "7",
		"--oem", "1",
		"-l", "eng",
		"-c", "tessedit_char_whitelist=0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("tesseract error: %v", err)
	}

	result := strings.TrimSpace(out.String())
	if result == "" {
		return "", fmt.Errorf("no text detected")
	}

	return result, nil
}
