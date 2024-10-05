package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

var (
	audioBuffer     bytes.Buffer
	isStreaming     bool
	streamStartTime time.Time
	currentFileName string
)

func main() {
	app := fiber.New()

	app.Post("/stream", handleStream)

	fmt.Println("Server is running on :4000")
	app.Listen(":4000")
}

func handleStream(c *fiber.Ctx) error {
	if c.Get("Content-Type") != "audio/mpeg" {
		return c.Status(400).SendString("Invalid Content-Type. Expected audio/mpeg")
	}

	chunk := c.Body()
	_, err := audioBuffer.Write(chunk)
	if err != nil {
		return c.Status(500).SendString("Failed to process audio chunk")
	}

	// Check if 10 seconds have passed or if this is the first chunk
	if !isStreaming || time.Since(streamStartTime) >= 10*time.Second {
		if err := saveBufferToFile(); err != nil {
			return c.Status(500).SendString("Failed to save audio data: " + err.Error())
		}
		audioBuffer.Reset()
		streamStartTime = time.Now()
		isStreaming = true
	}

	return c.SendStatus(200)
}

func saveBufferToFile() error {
	if audioBuffer.Len() == 0 {
		return nil // Nothing to save
	}

	currentFileName = generateFileName()
	filePath := filepath.Join(".", currentFileName) // Save in the current directory

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, &audioBuffer)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	fmt.Printf("Saved audio to %s\n", filePath)
	return nil
}

func generateFileName() string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("broadcast_%s.mp3", timestamp)
}
