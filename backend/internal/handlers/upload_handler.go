package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UploadHandler struct {
	uploadDir string
}

func NewUploadHandler() *UploadHandler {
	dir := filepath.Join(".", "uploads", "products")
	os.MkdirAll(dir, 0755)
	return &UploadHandler{uploadDir: dir}
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

	filename := uuid.New().String() + ext
	dstPath := filepath.Join(h.uploadDir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "failed to save file"})
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(dstPath)
		return c.Status(500).JSON(fiber.Map{"success": false, "error": "failed to write file"})
	}

	url := fmt.Sprintf("/uploads/products/%s", filename)
	return c.JSON(fiber.Map{"success": true, "data": fiber.Map{"url": url}})
}
