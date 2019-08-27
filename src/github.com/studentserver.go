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
	ID    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` // bson to marshall and unmarshal
	Name  string             `json:"name,omitempty" bson:"name,omitempty"`
	Email string             `json:"email,omitempty" bson:"email,omitempty"`
}

func CreateContactEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var contact Contact
	_ = json.NewDecoder(request.Body).Decode(&contact)
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

	err := collection.FindOne(ctx, Contact{ID: id}).Decode(&contact)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))

		// dat := collection.FindOne(ctx, Contact{ID: id})
		// dat.Decode(&contact)
		// if dat != nil {
		// response.WriteHeader(http.StatusInternalServerError)
		// http.Error(response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	json.NewEncoder(response).Encode(contact)
}

func DeleteContactendpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	var contact Contact
	var err error
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	collection := client.Database("contactlist").Collection("contactlist")
	d, err := collection.DeleteOne(ctx, Contact{ID: id})

	if d.DeletedCount == 0 {
		fmt.Println("hELLO_from error")
		fmt.Println(err)
		response.WriteHeader(http.StatusInternalServerError)
		// response.Write([]byte(`{ "message": Data Not Found"` + `" }`))
		http.Error(response, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		fmt.Println(response)
		return
	}
	json.NewEncoder(response).Encode(contact)
}
func main() {
	fmt.Println("Starting the application...")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/student", CreateContactEndpoint).Methods("POST")
	router.HandleFunc("/student", GetPeopleEndpoint).Methods("GET")
	router.HandleFunc("/student/{id}", GetContactEndpoint).Methods("GET")
	router.HandleFunc("/student/{id}", DeleteContactendpoint).Methods("DELETE")
	http.ListenAndServe(":3004", router)
}
