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
	audioBuffer bytes.Buffer
	isStreaming bool
)

func main() {
	app := fiber.New()

	app.Post("/start-stream", handleStartStream)
	app.Post("/stream", handleStream)
	app.Post("/end-stream", handleEndStream)

	fmt.Println("Server is running on :4000")
	app.Listen(":3000")
}

func handleStartStream(c *fiber.Ctx) error {
	if isStreaming {
		return c.Status(400).SendString("A stream is already in progress")
	}
	audioBuffer.Reset()
	isStreaming = true
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

	return c.SendString("Audio chunk received")
}

func handleEndStream(c *fiber.Ctx) error {
	if !isStreaming {
		return c.Status(400).SendString("No active stream to end")
	}

	isStreaming = false

	timestamp := time.Now().Format("20060102_150405")
	fileName := fmt.Sprintf("broadcast_%s.mp3", timestamp)

	file, err := os.Create(fileName)
	if err != nil {
		return c.Status(500).SendString("Failed to create file")
	}
	defer file.Close()

	_, err = io.Copy(file, &audioBuffer)
	if err != nil {
		return c.Status(500).SendString("Failed to save audio data")
	}

	audioBuffer.Reset()
	return c.SendString("Stream ended and saved to " + fileName)
}
