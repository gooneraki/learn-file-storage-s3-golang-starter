package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	const maxMemory = 10 << 20 // 10 MB
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")
	if mediaType == "" {
		respondWithError(w, http.StatusBadRequest, "Missing Content-Type for thumbnail", nil)
		return
	}

	// data, err := io.ReadAll(file)
	// if err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, "Error reading file", err)
	// 	return
	// }

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't find video", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Not authorized to update this video", nil)
		return
	}

	file_extension := ""
	if strings.Contains(mediaType, "png") {
		file_extension = "png"
	} else if strings.Contains(mediaType, "pdf") {
		file_extension = "pdf"

	} else if strings.Contains(mediaType, "mp4") {
		file_extension = "mp4"
	}

	if file_extension == "" {
		respondWithError(w, http.StatusInternalServerError, "couldn't get file extension", err)
		return
	}

	savepath := filepath.Join(cfg.assetsRoot, fmt.Sprintf("%s.%s", videoIDString, file_extension))
	newFile, err := os.Create(savepath)
	if err != nil {
		fmt.Println(savepath)
		respondWithError(w, http.StatusInternalServerError, "couldn't create file", err)
		return
	}

	if _, err := io.Copy(newFile, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't create file", err)
		return
	}

	// base64Encoded := base64.StdEncoding.EncodeToString(data)
	// base64DataURL := fmt.Sprintf("data:%s;base64,%s", mediaType, base64Encoded)
	// video.ThumbnailURL = &base64DataURL

	url := fmt.Sprintf("http://localhost:%s/assets/%s.%s", cfg.port, videoID, file_extension)
	video.ThumbnailURL = &url

	err = cfg.db.UpdateVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
