package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var rnd *renderer.Render
var collection *mongo.Collection

const (
	hostName       string = "localhost:27017"
	dbName         string = "just_todo"
	collectionName string = "todo"
	port           string = ":9000"
)

type (
	todoModel struct {
		ID        primitive.ObjectID `bson:"_id,omitempty"`
		Title     string             `bson:"title"`
		Completed bool               `bson:"completed"`
		CreatedAt time.Time          `bson:"createdat"`
	}
	todo struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		Completed bool      `json:"completed"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Connect Database
func init() {
	rnd = renderer.New()

	clientOptions := options.Client().ApplyURI("mongodb://" + hostName)
	client, err := mongo.Connect(context.Background(), clientOptions)
	checkErr(err)

	db := client.Database(dbName)
	collection = db.Collection(collectionName)
}

func fetchTodos(w http.ResponseWriter, r *http.Request) {
	todos := []todoModel{}

	// Fetch data from the database
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to fetch todo",
			"error":   err,
		})
		return
	}
	defer cursor.Close(context.Background())

	err = cursor.All(context.Background(), &todos)
	if err != nil {
		rnd.JSON(w, http.StatusProcessing, renderer.M{
			"message": "Failed to fetch todo",
			"error":   err,
		})
		return
	}

	todoList := []todo{}
	for _, i := range todos {
		todoList = append(todoList, todo{
			ID:        i.ID.Hex(),
			Title:     i.Title,
			Completed: i.Completed,
			CreatedAt: i.CreatedAt,
		})
	}

	// Render the response
	rnd.JSON(w, http.StatusOK, renderer.M{
		"data": todoList,
	})
}

func createTodo(w http.ResponseWriter, r *http.Request) {
	// Parse the request http body to extract or passing a reference to newTodo
	var newTodo todo
	err := json.NewDecoder(r.Body).Decode(&newTodo)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if newTodo.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "The title is required",
		})
		return
	}

	// Create a todoModel instance from the newTodo data
	todoModel := todoModel{
		ID:        primitive.NewObjectID(),
		Title:     newTodo.Title,
		Completed: false,
		CreatedAt: time.Now(),
	}

	// Insert the todoModel into the MongoDB collection
	_, err = collection.InsertOne(context.Background(), todoModel)
	if err != nil {
		log.Printf("Error inserting todo: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return a success response
	rnd.JSON(w, http.StatusCreated, renderer.M{
		"message": "Todo created successfully",
		"todo_id": todoModel.ID.Hex(),
	})
}

func updateTodo(w http.ResponseWriter, r *http.Request) {
	// Extract todo ID from URL parameters
	todoID := chi.URLParam(r, "id")

	// Validate the todo ID
	if !primitive.IsValidObjectID(todoID) {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	// Convert the string ID to ObjectId
	objID, _ := primitive.ObjectIDFromHex(todoID)

	// Parse the request body to get updated todo data
	var updatedTodo todo
	err := json.NewDecoder(r.Body).Decode(&updatedTodo)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if updatedTodo.Title == "" {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"	message": "The title is required",
		})
		return
	}

	// Update the todo in the MongoDB collection
	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"title":     updatedTodo.Title,
			"completed": true,
		},
	}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Error updating todo: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return a success response
	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo updated successfully",
		"todo_id": todoID,
	})
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	// Extract todo ID from URL parameters
	todoID := chi.URLParam(r, "id")

	// Validate the todo ID
	if !primitive.IsValidObjectID(todoID) {
		http.Error(w, "Invalid todo ID", http.StatusBadRequest)
		return
	}

	// Convert the string ID to ObjectId
	objID, _ := primitive.ObjectIDFromHex(todoID)

	// Remove the todo from the MongoDB collection
	_, err := collection.DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		log.Printf("Error deleting todo: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Return a success response
	rnd.JSON(w, http.StatusOK, renderer.M{
		"message": "Todo deleted successfully",
		"todo_id": todoID,
	})
}

func todoHandlers() http.Handler {
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/", fetchTodos)
		r.Post("/", createTodo)
		r.Put("/{id}", updateTodo)
		r.Delete("/{id}", deleteTodo)
	})
	return rg
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := rnd.Template(w, http.StatusOK, []string{"frontend/home.html"}, nil)
	checkErr(err)
}

func main() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", homeHandler)
	r.Mount("/todo", todoHandlers())

	server := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Println("Listening on Port: ", port)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
		}
	}()

	<-stopChan
	log.Println("shutting down server....")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	server.Shutdown(ctx)

	defer cancel()
	log.Println("Server stopped")
}

