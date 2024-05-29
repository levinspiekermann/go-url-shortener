package server

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func (s *FiberServer) RegisterFiberRoutes() {
	s.App.Get("/:shortened", s.ShortenedUrlHandler)
	s.App.Post("/shorten", s.ShortenUrlHandler)
	s.App.Get("/health", s.healthHandler)

}

func (s *FiberServer) ShortenedUrlHandler(c *fiber.Ctx) error {
	data, err := s.db.GetShortenedUrlByShortURL(c.Params("shortened"))
	log.Default().Println(data, c.Params("shortened"))

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Not Found",
		})
	}

	return c.Redirect(data, fiber.StatusMovedPermanently)
}

func (s *FiberServer) ShortenUrlHandler(c *fiber.Ctx) error {
	var body struct {
		URL string `json:"url"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON",
		})
	}

	shortenedURL, err := s.db.CreateShortenedUrl(body.URL)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
		})
	}

	return c.JSON(fiber.Map{
		"shortened": shortenedURL.ShortURL,
	})
}

func (s *FiberServer) healthHandler(c *fiber.Ctx) error {
	return c.JSON(s.db.Health())
}
