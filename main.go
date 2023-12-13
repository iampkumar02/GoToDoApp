package main

import(
	"fmt"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
	"context"
	"os"
	"os/signal"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var rnd *renderer.Render
var db *mgo.Database

const (
	hostName			string = "localhost:12000"
	dbName				string = "just_todo"
	collectionName		string = "todo"
	port				string = ":9000"
)


type(
	todoModel struct{
		ID			string `bson:"_id,omitempty"`
		Title		string `bson:"title"`
		Completed	bool `bson:"completed"`
		CreatedAt 	time.Time `bson:"createdat"`
	}
	todo struct{
		ID			string `json:"id"`
		Title		string `json:"title"`
		Completed	string `json:"completed"`
		CreatedAt	time.Time `json:"created_at"`
	}
)


func checkErr(err error){
	if err!=nil {
		log.Fatal(err)
	}
}

func init(){
	rnd = renderer.New()
	sess,err:=mgo.Dial(hostName)
	checkErr(err)
	sess.SetMode(mgo.Monotonic,true)
	db = sess.DB(dbName)

}

func todoHandlers() http.Handler{
	rg := chi.NewRouter()
	rg.Group(func(r chi.Router) {
		r.Get("/",fetchTodos)
		r.Post("/",createTodo)
		r.Put("/{id}",updateTodo)
		r.Delete("/{id}",deleteTodo)
	})
	return rg
}

func main(){
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/",homeHandler)
	r.Mount("/todo",todoHandlers())

	server := &http.Server{
		Addr: port,
		Handler: r,
		ReadTimeout: 60*time.Second,
		WriteTimeout: 60*time.Second,
		IdleTimeout: 60*time.Second,
	}

	go func(){
		log.Println("Listening on Port: ",port)
		if err:=server.ListenAndServe(); err!=nil{
			log.Printf("listen: %s\n",err)
		}
	}()
}