package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)


type Server struct {
	client *mongo.Client
}

func NewServer(c *mongo.Client) *Server{
	return &Server{
		client: c,
	}
}

func (s *Server) handleGetAllFacts(w http.ResponseWriter, r *http.Request) {
	coll := s.client.Database("dogfact").Collection("facts")

	query := bson.M{}

	cursor, err := coll.Find(context.TODO(), query)

	if err != nil {
		log.Fatal(err)
	}

	results := []bson.M{}
	// check for errors in the conversion
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

type CatFactWorker struct {
	client *mongo.Client
}

func NewDogFactWorker(c *mongo.Client) *CatFactWorker {

	return &CatFactWorker{
		client: c,
	}
}

func (cfw *CatFactWorker) start() error {
	coll := cfw.client.	Database("dogfact").Collection("facts")
	ticker := time.NewTicker(2 * time.Second)

	for {
		response, err := http.Get("https://catfact.ninja/fact")

		if err != nil {
			return err
		}

		var catFact bson.M //map[string]any or map[string]interface{}

		if err := json.NewDecoder(response.Body).Decode(&catFact); err != nil {
			return err
		}

		_, err = coll.InsertOne(context.TODO(), catFact)

		if err != nil {
			return err
		}

		// fmt.Println(catFact)
		<-ticker.C
	}
}

func main() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))

	if err != nil {
		panic(err)
	}

	// fmt.Println(client)

	worker := NewDogFactWorker(client)
	go worker.start()

	server := NewServer(client)

	http.HandleFunc("/facts", server.handleGetAllFacts)
	http.ListenAndServe(":3000", nil)
}
