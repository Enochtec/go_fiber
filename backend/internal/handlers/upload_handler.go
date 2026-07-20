package handlers

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type UploadHandler struct{}

func NewUploadHandler() *UploadHandler {
	return &UploadHandler{}
}

var allowedExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
}

var allowedMimes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/webp": ".webp",
}

const maxSize = 5 << 20

func (h *UploadHandler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "no file provided"})
	}
	if file.Size > maxSize {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "file too large (max 5MB)"})
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "unsupported format (PNG, JPG, WEBP only)"})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "failed to read file"})
	}
	defer src.Close()

	buf := make([]byte, 512)
	if _, err := src.Read(buf); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "corrupted file"})
	}
	src.Seek(0, 0)

	mimeType := http.DetectContentType(buf)
	if _, ok := allowedMimes[mimeType]; !ok {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "invalid image content"})
	}

	url, err := uploadToCloudinary(src, file.Filename)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "cloudinary upload failed: " + err.Error()})
	}
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"url": url}})
}

func uploadToCloudinary(file io.Reader, filename string) (string, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey    := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return "", fmt.Errorf("CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY and CLOUDINARY_API_SECRET must be set")
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	folder    := "pos/products"

	// Build signature: SHA-1("folder=...&timestamp=..." + apiSecret)
	sigPayload := fmt.Sprintf("folder=%s&timestamp=%s%s", folder, timestamp, apiSecret)
	h := sha1.New()
	h.Write([]byte(sigPayload))
	sig := fmt.Sprintf("%x", h.Sum(nil))

	// Build multipart body
	body   := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}
	writer.WriteField("api_key", apiKey)
	writer.WriteField("timestamp", timestamp)
	writer.WriteField("signature", sig)
	writer.WriteField("folder", folder)
	writer.Close()

	endpoint := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", cloudName)
	req, err  := http.NewRequest("POST", endpoint, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if secureURL, ok := result["secure_url"].(string); ok {
		return secureURL, nil
	}
	if errMap, ok := result["error"].(map[string]interface{}); ok {
		if msg, ok := errMap["message"].(string); ok {
			return "", fmt.Errorf(msg)
		}
	}
	return "", fmt.Errorf("unexpected cloudinary response")
}
