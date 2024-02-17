package main

import (
	"log"

	"github.com/Yadier01/golangbackendlearn/db"
	"github.com/Yadier01/golangbackendlearn/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()

	collection, err := db.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	app.Use(cors.New())
	todoHandler := &handlers.TodosHandler{Collection: collection}

	app.Get("/", todoHandler.GetTodos)
	app.Post("/", todoHandler.PostTodo)
	app.Delete("/:id", todoHandler.DeleteTodo)
	app.Patch("/:id", todoHandler.PatchTodo)

	app.Static("/image", "./public/random.jpg")

	log.Fatal(app.Listen(":3002"))

}
