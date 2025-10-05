package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	const uploadLimit = 1 << 30
	// r.read
	http.MaxBytesReader(w, r.Body, uploadLimit)

	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid video ID", err)
		return
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't get bearer token", err)
		return
	}

	userId, err := auth.ValidateJWT(tokenString, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't get bearer token", err)
		return
	}

	videoDb, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't get video", err)
		return
	}

	if videoDb.UserID != userId {
		respondWithError(w, http.StatusUnauthorized, "you are not authorised to change this video", err)
		return

	}

	file, _, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't get file and header", err)
		return
	}

	defer file.Close()

	mediaType, _, err := mime.ParseMediaType("video/mp4")

	f, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't create temp file", err)
		return
	}
	defer os.Remove(f.Name()) // clean up
	defer f.Close()

	if _, err = io.Copy(f, file); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error saving file", err)
		return
	}

	f.Seek(0, io.SeekStart)

	key := getAssetPath(mediaType)

	_, err = cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         aws.String(key),
		Body:        f,
		ContentType: aws.String("video/mp4"),
	})

	newUrl := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, key)

	err = cfg.db.UpdateVideo(database.Video{
		ID:                videoID,
		CreatedAt:         videoDb.CreatedAt,
		UpdatedAt:         time.Now(),
		ThumbnailURL:      videoDb.ThumbnailURL,
		VideoURL:          &newUrl,
		CreateVideoParams: videoDb.CreateVideoParams,
	})

	// if _, err := f.Write([]byte("content")); err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, "couldn't write file contents ", err)
	// 	return
	// }
	// if err := f.Close(); err != nil {
	// 	respondWithError(w, http.StatusInternalServerError, "couldn't close the file ", err)
	// 	return
	// }

}
