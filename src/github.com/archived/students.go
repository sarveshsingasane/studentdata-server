package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Contact struct {
	ID    primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name  string             `json:"name,omitempty" bson:"name,omitempty"`
	Email string             `json:"email,omitempty" bson:"email,omitempty"`
}

func CreateContactEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var contact Contact
	json.NewDecoder(request.Body).Decode(&contact)
	collection := client.Database("contactlist").Collection("contactlist")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	result, err := collection.InsertOne(ctx, contact)
	contact.ID = result.InsertedID.(primitive.ObjectID)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(contact)
}

func GetPeopleEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var people []Contact
	collection := client.Database("contactlist").Collection("contactlist")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var contact Contact
		cursor.Decode(&contact)
		people = append(people, contact)
	}
	err = cursor.Err()
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(people)
}

func GetContactEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var contact Contact
	collection := client.Database("contactlist").Collection("contactlist")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	idDoc := bson.D{{"_id", id}}
	err := collection.FindOne(ctx, idDoc).Decode(&contact)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))

		return
	}
	json.NewEncoder(response).Encode(contact)
}

func DeleteContactendpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	// var contact Contact
	// var err error
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	collection := client.Database("contactlist").Collection("contactlist")
	idDoc := bson.D{{"_id", id}}
	d, _ := collection.DeleteOne(ctx, idDoc)
	if d.DeletedCount == 0 {
		response.WriteHeader(http.StatusInternalServerError)
		http.Error(response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else {
		json.NewEncoder(response).Encode(d)
	}
}
func main() {
	fmt.Println("Starting the server...")

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/students", CreateContactEndpoint).Methods("POST")
	router.HandleFunc("/students", GetPeopleEndpoint).Methods("GET")
	router.HandleFunc("/students/{id}", GetContactEndpoint).Methods("GET")
	router.HandleFunc("/students/{id}", DeleteContactendpoint).Methods("DELETE")
	http.ListenAndServe(":3005", router)
}
