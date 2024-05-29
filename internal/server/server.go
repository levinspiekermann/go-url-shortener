package server

import (
	"github.com/gofiber/fiber/v2"

	"go-url-shortener/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "go-url-shortener",
			AppName:      "go-url-shortener",
		}),

		db: database.New(),
	}

	return server
}
