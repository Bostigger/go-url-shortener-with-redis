package routes

import (
	"github.com/go-url-shortener/database"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

func ResolveUrl(c *fiber.Ctx) error {
	url := c.Params("url")
	r := database.CreateClient(0)
	defer r.Close()

	value, err := r.Get(database.Ctx, url).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Url was not found"})
	} else if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Database connection error"})
	}
	rInr := database.CreateClient(1)
	defer rInr.Close()

	_ = rInr.Incr(database.Ctx, "counter")
	return c.Redirect(value, 301)
}
