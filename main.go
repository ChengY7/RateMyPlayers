package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

type Player struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Position  string `json:"position"`
	Team      string `json:"team"`
	Age       int    `json:"age"`
	Active    bool   `json:"active"`
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is not set in the environment")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to the database!")

	app := fiber.New()

	app.Get("/api/players", func(c *fiber.Ctx) error {
		return getPlayers(c, db)
	})
	app.Post("/api/players", func(c *fiber.Ctx) error {
		return createPlayer(c, db)
	})
	app.Patch("/api/players/:id", func(c *fiber.Ctx) error {
		return updatePlayer(c, db)
	})
	app.Delete("/api/players/:id", func(c *fiber.Ctx) error {
		return deletePlayer(c, db)
	})

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "4000"
	}

	log.Fatal(app.Listen(":" + PORT))
}

func getPlayers(c *fiber.Ctx, db *sql.DB) error {
	rows, err := db.Query("SELECT id, first_name, last_name, position, team, age, active FROM players")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch players"})
	}
	defer rows.Close()

	var players []Player
	for rows.Next() {
		var player Player
		if err := rows.Scan(&player.ID, &player.FirstName, &player.LastName, &player.Position, &player.Team, &player.Age, &player.Active); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to scan player"})
		}
		players = append(players, player)
	}

	return c.JSON(players)
}

func createPlayer(c *fiber.Ctx, db *sql.DB) error {
	player := new(Player)
	if err := c.BodyParser(player); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if player.FirstName == "" || player.LastName == "" || player.Position == "" || player.Team == "" || player.Age == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Missing required fields"})
	}

	query := "INSERT INTO players (first_name, last_name, position, team, age, active) VALUES (?, ?, ?, ?, ?, ?)"
	result, err := db.Exec(query, player.FirstName, player.LastName, player.Position, player.Team, player.Age, player.Active)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create player"})
	}

	lastInsertID, _ := result.LastInsertId()
	player.ID = int(lastInsertID)

	return c.Status(201).JSON(player)
}

func updatePlayer(c *fiber.Ctx, db *sql.DB) error {
	id := c.Params("id")

	type UpdatePlayerRequest struct {
		FirstName *string `json:"first_name"`
		LastName  *string `json:"last_name"`
		Position  *string `json:"position"`
		Team      *string `json:"team"`
		Age       *int    `json:"age"`
		Active    *bool   `json:"active"`
	}

	req := new(UpdatePlayerRequest)
	if err := c.BodyParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}
	fmt.Println(req)
	fields := map[string]interface{}{}

	if req.FirstName != nil {
		fields["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		fields["last_name"] = *req.LastName
	}
	if req.Position != nil {
		fields["position"] = *req.Position
	}
	if req.Team != nil {
		fields["team"] = *req.Team
	}
	if req.Age != nil {
		fields["age"] = *req.Age
	}
	if req.Active != nil {
		fields["active"] = *req.Active
	}

	if len(fields) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No fields to update"})
	}

	query := "UPDATE players SET "
	args := []interface{}{}
	for column, value := range fields {
		query += column + " = ?, "
		args = append(args, value)
	}
	query = query[:len(query)-2] + " WHERE id = ?" // Remove trailing comma and add WHERE clause
	args = append(args, id)

	_, err := db.Exec(query, args...)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update player", "details": err.Error()})
	}

	return c.Status(200).JSON(fiber.Map{"message": "Player updated successfully"})
}

func deletePlayer(c *fiber.Ctx, db *sql.DB) error {
	id := c.Params("id")

	query := "DELETE FROM players WHERE id = ?"
	_, err := db.Exec(query, id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete player"})
	}

	return c.Status(200).JSON(fiber.Map{"message": "Player deleted successfully"})
}
