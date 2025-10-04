package main

import (
	"fmt"
	"mime"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetPath(videoID uuid.UUID, mediaType string) (string, error) {
	ext, err := mediaTypeToExt(mediaType)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", videoID, ext), nil
}

func (cfg apiConfig) getAssetDiskPath(assetPath string) string {
	return filepath.Join(cfg.assetsRoot, assetPath)
}

func (cfg apiConfig) getAssetURL(assetPath string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetPath)
}

func mediaTypeToExt(mediaType string) (string, error) {

	mediatype, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return "", err
	}

	if mediatype != "image/jpeg" || mediaType != "image/png" {
		return "", fmt.Errorf("unsupported media type %s", mediaType)
	}

	return mediaType, nil

}
