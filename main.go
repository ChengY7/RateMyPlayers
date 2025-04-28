package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

type Player struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

func main() {
	fmt.Println("Hello, World!!!!!")
	app := fiber.New()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	PORT := os.Getenv("PORT")

	players := []Player{}

	app.Get("/api/players", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(players)
	})

	app.Post("/api/players", func(c *fiber.Ctx) error {
		player := &Player{}

		if err := c.BodyParser(player); err != nil {
			return err
		}

		if player.Name == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Name is required"})
		}

		player.ID = len(players) + 1
		players = append(players, *player)
		return c.Status(201).JSON(player)
	})

	app.Patch("/api/players/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		for i, player := range players {
			if fmt.Sprintf("%d", player.ID) == id {
				players[i].Active = true
				return c.Status(200).JSON(players[i])
			}
		}

		return c.Status(404).JSON(fiber.Map{"error": "Player not found"})
	})

	app.Delete("/api/players/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		for i, player := range players {
			if fmt.Sprintf("%d", player.ID) == id {
				players = append(players[:i], players[i+1:]...)
				return c.Status(200).JSON(fiber.Map{"message": "Player deleted"})
			}
		}

		return c.Status(404).JSON(fiber.Map{"error": "Player not found"})
	})

	log.Fatal(app.Listen(":" + PORT))
}
