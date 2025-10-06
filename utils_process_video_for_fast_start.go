package main

import (
	"fmt"
	"os/exec"
)

func processVideoForFastStart(filePath string) (string, error) {
	outpath := filePath + ".processing"

	cmd := exec.Command("ffmpeg",
		"-i", filePath,
		"-c", "copy",
		"-movflags", "faststart",
		"-f", "mp4",
		outpath,
	)

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("ffmpeg error: %v", err)
	}

	return outpath, nil
}
