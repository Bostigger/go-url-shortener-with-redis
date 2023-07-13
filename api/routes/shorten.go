package routes

import (
	"github.com/asaskevich/govalidator"
	"github.com/go-url-shortener/database"
	"github.com/go-url-shortener/helpers"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"os"
	"strconv"
	"time"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"customShort"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL             string        `json:"url"`
	CustomShort     string        `json:"customShort"`
	Expiry          time.Duration `json:"expiry"`
	XRateRemaining  int           `json:"rate_limit"`
	XRateLimitReset time.Duration `json:"rate_limit_reset"`
}

func ShortenUrl(c *fiber.Ctx) error {
	body := new(request)
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}

	//implement rate limiting
	r2 := database.CreateClient(1)
	defer r2.Close()
	value, err := r2.Get(database.Ctx, c.IP()).Result()
	if err == redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("APP_QUOTA"), time.Second*30).Err()
	} else {
		value, _ = r2.Get(database.Ctx, c.IP()).Result()
		valInt, _ := strconv.Atoi(value)
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Limit exceeded", "rate_time_rest": limit / time.Nanosecond / time.Minute})
		}
	}

	//check if its actual url
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "check url"})
	}

	//check for domain err
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "check url"})
	}

	//enforce https, ssl
	body.URL = helpers.EnforceHTTP(body.URL)

	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}
	r := database.CreateClient(0)
	defer r.Close()

	val, _ := r.Get(database.Ctx, id).Result()
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Your shor is already used"})
	}
	if body.Expiry == 0 {
		body.Expiry = 24
	}
	err = r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to connect to the server"})
	}

	resp := response{
		URL:             body.URL,
		CustomShort:     "",
		Expiry:          body.Expiry,
		XRateRemaining:  10,
		XRateLimitReset: 30,
	}
	r2.Decr(database.Ctx, c.IP())

	val, _ = r2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateLimitReset = ttl / time.Nanosecond / time.Minute
	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": resp})
}
