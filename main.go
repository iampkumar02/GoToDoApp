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
		ID			bson.ObjectId `bson:"_id,omitempty"`
		Title		string `bson:"title"`
		Completed	bool `bson:"completed"`
		CreatedAt 	time.Time `bson:"createdat"`
	}
	todo struct{
		ID			string `json:"id"`
		Title		string `json:"title"`
		Completed	bool `json:"completed"`
		CreatedAt	time.Time `json:"created_at"`
	}
)


func checkErr(err error){
	if err!=nil {
		log.Fatal(err)
	}
}


// Connect Database
func init(){
	rnd = renderer.New()
	sess,err:=mgo.Dial(hostName)
	checkErr(err)
	sess.SetMode(mgo.Monotonic,true)
	db = sess.DB(dbName)

}

func fetchTodos(w http.ResponseWriter, r *http.Request){
	todos := []todoModel{}

	if err:=db.C(collectionName).Find(bson.M{}).All(&todos);err!=nil{
		rnd.JSON(w,http.StatusProcessing,renderer.M{
			"message":"Failed to fetch todo",
			"error":err,
		})
		return
	}

	todoList := []todo{}
	for _,i := range todos{
		todoList = append(todoList,todo{
			ID: i.ID.Hex(),
			Title: i.Title,
			Completed: i.Completed,
			CreatedAt: i.CreatedAt,
		})
	}

	rnd.JSON(w, http.StatusOK,renderer.M{
		"data": todoList,
	})

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

func homeHandler(w http.ResponseWriter, r *http.Request){
	err:=rnd.Template(w, http.StatusOK, []string{"static/home.tpl"},nil)
	checkErr(err)
}


func main(){
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan,os.Interrupt)
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

	<-stopChan
	log.Println("shutting down server....")
	ctx, cancel := context.WithTimeout(context.Background(),5*time.Second)
	server.Shutdown(ctx)

	defer cancel(
		// log.Println("Server stopped!")
	)
}