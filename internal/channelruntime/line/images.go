package line

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/quailyquaily/mistermorph/internal/telegramutil"
	"github.com/quailyquaily/mistermorph/llm"
)

const (
	lineLLMMaxImages     = 3
	lineLLMMaxImageBytes = int64(5 * 1024 * 1024)
)

func buildLineHistoryMessage(content string, model string, imagePaths []string, logger *slog.Logger) (llm.Message, error) {
	msg := llm.Message{Role: "user", Content: content}
	if !llm.ModelSupportsImageParts(model) || len(imagePaths) == 0 {
		return msg, nil
	}
	parts := make([]llm.Part, 0, 1+min(len(imagePaths), lineLLMMaxImages))
	if strings.TrimSpace(content) != "" {
		parts = append(parts, llm.Part{Type: llm.PartTypeText, Text: content})
	}

	seen := make(map[string]bool, len(imagePaths))
	imageCount := 0
	for _, rawPath := range imagePaths {
		if imageCount >= lineLLMMaxImages {
			break
		}
		path := strings.TrimSpace(rawPath)
		if path == "" || seen[path] {
			continue
		}
		seen[path] = true

		info, err := os.Stat(path)
		if err != nil {
			if logger != nil {
				logger.Warn("line_image_part_skip", "path", path, "error", err.Error())
			}
			continue
		}
		if info.Size() <= 0 {
			continue
		}
		if info.Size() > lineLLMMaxImageBytes {
			return llm.Message{}, fmt.Errorf("图片太大: %s (%d bytes > %d bytes)", filepath.Base(path), info.Size(), lineLLMMaxImageBytes)
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			if logger != nil {
				logger.Warn("line_image_part_read_error", "path", path, "error", err.Error())
			}
			continue
		}
		mimeType := lineImageMIMEType(path)
		if !isLineSupportedUploadImageMIME(mimeType) {
			if logger != nil {
				logger.Warn("line_image_part_skip_unsupported_format", "path", path, "mime_type", mimeType)
			}
			continue
		}

		parts = append(parts, llm.Part{
			Type:       llm.PartTypeImageBase64,
			MIMEType:   mimeType,
			DataBase64: base64.StdEncoding.EncodeToString(raw),
		})
		imageCount++
	}
	if imageCount == 0 {
		return msg, nil
	}
	msg.Parts = parts
	return msg, nil
}

func downloadLineImageToCache(ctx context.Context, api *lineAPI, cacheDir string, messageID string, maxBytes int64) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if api == nil {
		return "", fmt.Errorf("line api is not initialized")
	}
	cacheDir = strings.TrimSpace(cacheDir)
	if cacheDir == "" {
		return "", fmt.Errorf("line image cache dir is required")
	}
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return "", fmt.Errorf("line message id is required")
	}
	if maxBytes <= 0 {
		return "", fmt.Errorf("line image max bytes must be positive")
	}
	if err := telegramutil.EnsureSecureCacheDir(cacheDir); err != nil {
		return "", err
	}

	raw, mimeType, err := api.messageContent(ctx, messageID, maxBytes)
	if err != nil {
		return "", err
	}
	mimeType = lineNormalizeMIMEType(mimeType)
	if !isLineSupportedUploadImageMIME(mimeType) {
		return "", fmt.Errorf("line image format is not supported: %s", mimeType)
	}
	ext := lineImageExtFromMIMEType(mimeType)
	if ext == "" {
		return "", fmt.Errorf("line image extension is not supported: %s", mimeType)
	}

	pattern := "line_" + sanitizeLineFileToken(messageID) + "_*" + ext
	tmp, err := os.CreateTemp(cacheDir, pattern)
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return tmpPath, nil
}

func ensureLineSecureChildDir(parentDir, childDir string) error {
	parentDir = strings.TrimSpace(parentDir)
	childDir = strings.TrimSpace(childDir)
	if parentDir == "" || childDir == "" {
		return fmt.Errorf("missing parent/child dir")
	}
	parentAbs, err := filepath.Abs(parentDir)
	if err != nil {
		return err
	}
	childAbs, err := filepath.Abs(childDir)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(parentAbs, childAbs)
	if err != nil {
		return err
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("child dir is not under parent dir: %s", childAbs)
	}
	return telegramutil.EnsureSecureCacheDir(childAbs)
}

func lineNormalizeMIMEType(mimeType string) string {
	mimeType = strings.TrimSpace(strings.ToLower(mimeType))
	if idx := strings.Index(mimeType, ";"); idx >= 0 {
		mimeType = strings.TrimSpace(mimeType[:idx])
	}
	return mimeType
}

func lineImageExtFromMIMEType(mimeType string) string {
	switch lineNormalizeMIMEType(mimeType) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

func lineImageMIMEType(path string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(path)))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return ""
	}
}

func isLineSupportedUploadImageMIME(mimeType string) bool {
	switch lineNormalizeMIMEType(mimeType) {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

func sanitizeLineFileToken(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "img"
	}
	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		return "img"
	}
	return out
}
