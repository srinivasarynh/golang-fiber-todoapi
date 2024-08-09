package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Task      string             `json:"task"`
	Completed bool               `json:"completed"`
}

const dbName = "resolution"
const colName = "todos"

var collection *mongo.Collection

func main() {
	fmt.Println("main.go file")

	enverr := godotenv.Load(".env")

	if enverr != nil {
		log.Fatal("error in loading env file")
	}

	PORT := os.Getenv("PORT")
	MONGO_URI := os.Getenv("MONGO_URI")

	clientOptions := options.Client().ApplyURI(MONGO_URI)
	client, mongoerr := mongo.Connect(context.Background(), clientOptions)
	defer client.Disconnect(context.Background())

	if mongoerr != nil {
		log.Fatal("mongodb connection failed", mongoerr)
	}

	collection = client.Database(dbName).Collection(colName)
	fmt.Println("db connected")

	app := fiber.New()

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodo)
	app.Patch("/api/todos/:id", updateTodo)
	app.Delete("/api/todos/:id", deleteTodo)

	apperr := app.Listen(":" + PORT)
	if apperr != nil {
		log.Fatal(apperr)
	}
}

func getTodos(c *fiber.Ctx) error {
	var todos []Todo

	col, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		return err
	}

	for col.Next(context.Background()) {
		var todo Todo
		if err := col.Decode(&todo); err != nil {
			return err
		}
		todos = append(todos, todo)
	}

	return c.Status(200).JSON(todos)

}

func createTodo(c *fiber.Ctx) error {
	todo := new(Todo)

	if err := c.BodyParser(todo); err != nil {
		return err
	}

	if todo.Task == "" {
		return c.Status(400).JSON(fiber.Map{
			"message": "task must have title",
		})
	}

	insertResult, err := collection.InsertOne(context.Background(), todo)

	if err != nil {
		return err
	}

	todo.ID = insertResult.InsertedID.(primitive.ObjectID)
	return c.Status(200).JSON(todo)
}

func updateTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "invalied id",
		})
	}

	filter := bson.M{"_id": objectId}
	update := bson.M{"$set": bson.M{"completed": true}}
	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "update success",
	})
}

func deleteTodo(c *fiber.Ctx) error {
	id := c.Params("id")
	objectId, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"message": "invalied id",
		})
	}

	filter := bson.M{"_id": objectId}
	_, err = collection.DeleteOne(context.Background(), filter)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "delete success",
	})
}
