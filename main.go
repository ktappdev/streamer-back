package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
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

	app.Post("/start-stream", handleStartStream)
	app.Post("/stream", handleStream)
	app.Post("/end-stream", handleEndStream)
	app.Post("/hi", hi)

	fmt.Println("Server is running on :4000")
	app.Listen(":4000")
}

func hi(c *fiber.Ctx) error {
	return c.SendString("Hello, World!")
}

func handleStartStream(c *fiber.Ctx) error {
	if isStreaming {
		return c.Status(400).SendString("A stream is already in progress")
	}
	audioBuffer.Reset()
	isStreaming = true
	streamStartTime = time.Now()
	currentFileName = generateFileName()
	return c.SendString("Stream started")
}

func handleStream(c *fiber.Ctx) error {
	if !isStreaming {
		return c.Status(400).SendString("No active stream. Call /start-stream first")
	}

	if c.Get("Content-Type") != "audio/mpeg" {
		return c.Status(400).SendString("Invalid Content-Type. Expected audio/mpeg")
	}

	chunk := c.Body()
	_, err := audioBuffer.Write(chunk)
	if err != nil {
		return c.Status(500).SendString("Failed to process audio chunk")
	}

	// Check if 10 seconds have passed
	if time.Since(streamStartTime) >= 10*time.Second {
		err := saveBufferToFile()
		if err != nil {
			return c.Status(500).SendString("Failed to save audio data: " + err.Error())
		}
		audioBuffer.Reset()
		streamStartTime = time.Now()
		currentFileName = generateFileName()
	}

	return c.SendString("Audio chunk received")
}

func handleEndStream(c *fiber.Ctx) error {
	if !isStreaming {
		return c.Status(400).SendString("No active stream to end")
	}

	err := saveBufferToFile()
	if err != nil {
		return c.Status(500).SendString("Failed to save final audio data: " + err.Error())
	}

	isStreaming = false
	audioBuffer.Reset()
	return c.SendString("Stream ended and saved to " + currentFileName)
}

func saveBufferToFile() error {
	if audioBuffer.Len() == 0 {
		return nil // Nothing to save
	}

	file, err := os.Create(currentFileName)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, &audioBuffer)
	if err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	fmt.Printf("Saved audio to %s\n", currentFileName)
	return nil
}

func generateFileName() string {
	timestamp := time.Now().Format("20060102_150405")
	return fmt.Sprintf("broadcast_%s.mp3", timestamp)
}
