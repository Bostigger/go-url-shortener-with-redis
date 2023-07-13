package main

import (
	"fmt"
	"github.com/go-url-shortener/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveUrl)
	app.Post("/api/v1", routes.ShortenUrl)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("URL Shortener")

	app := fiber.New()
	app.Use(logger.New())
	setupRoutes(app)

	err = app.Listen(os.Getenv("APP_PORT"))
	if err != nil {
		return
	}
}
