package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todos struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Description string             `json:"description"`
}
type TodosHandler struct {
	Collection *mongo.Collection
}

func (h *TodosHandler) GetTodos(c *fiber.Ctx) error {
	var result []*Todos
	//future me: c.Context will close connection if the user leaves before the request is finished
	ctx := c.Context()
	cur, err := h.Collection.Find(ctx, bson.D{{}}, options.Find().SetLimit(10))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("An error occurred")
	}
	//for future me, defer means that the function will be called at the end of the function
	// so this is jsut closing the context when we return or a panic occurs
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var elem Todos
		err := cur.Decode(&elem)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("An error occurred")
		}
		result = append(result, &elem)
	}
	return c.JSON(result)
}

func (h *TodosHandler) PostTodo(c *fiber.Ctx) error {
	todo := new(Todos)
	ctx := c.Context()
	if err := c.BodyParser(todo); err != nil {
		return c.Status(400).SendString("An error occurred")
	}
	var existingTodo Todos
	err := h.Collection.FindOne(ctx, bson.M{"description": todo.Description}).Decode(&existingTodo)
	// for future me: checks if erorr occured during db operation
	if err != nil && err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusInternalServerError).SendString("An error occurred")
	}

	// for future me: checks if the todo already exists
	if err != mongo.ErrNoDocuments {
		return c.Status(fiber.StatusBadRequest).SendString("Todo already exists")
	}

	insertResult, err := h.Collection.InsertOne(ctx, todo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("An error occurred while inserting the todo")
	}

	return c.JSON(fiber.Map{"id": insertResult.InsertedID})
}

func (h *TodosHandler) PatchTodo(c *fiber.Ctx) error {
	objID, err := primitive.ObjectIDFromHex(c.Params("id"))

	ctx := c.Context()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}

	todo := new(Todos)

	if err := c.BodyParser(todo); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Error parsing request body")
	}

	updateResult := h.Collection.FindOneAndUpdate(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{"description": todo.Description},
	})

	if err := updateResult.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return c.Status(fiber.StatusNotFound).SendString("No document found with the given ID")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("An error occurred while updating the document")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"id": objID, "succes": "Todo updated successfully", "newDescription": todo.Description,
	})

}

func (h *TodosHandler) DeleteTodo(c *fiber.Ctx) error {
	objID, err := primitive.ObjectIDFromHex(c.Params("id"))
	ctx := c.Context()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
	}
	res := h.Collection.FindOneAndDelete(ctx, bson.M{"_id": objID})
	if res.Err() != nil {
		return c.Status(fiber.StatusNotFound).SendString("No document found with the given ID")
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"description": "Todo deleted successfully", "id": objID,
	})
}
