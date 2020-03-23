package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	soajsgo "github.com/soajs/soajs.golang"
)

// Response example api response
type Response struct {
	Message string `json:"message"`
}

// Heartbeat heartbeat route handler
func Heartbeat(w http.ResponseWriter, r *http.Request) {
	resp := Response{}
	resp.Message = fmt.Sprintf("heartbeat")
	respJSON, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(respJSON)
}

func hello(w http.ResponseWriter, r *http.Request) {
	soajs := r.Context().Value(soajsgo.SoajsKey).(soajsgo.ContextData)
	respJSONSOA, err := json.Marshal(soajs)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(respJSONSOA)
}

//main function
func main() {
	router := mux.NewRouter()

	jsonFile, err := os.Open("soa.json")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully Opened soa.json")
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var soaConfig soajsgo.Config
	err = json.Unmarshal(byteValue, &soaConfig)
	if err != nil {
		log.Fatal(err)
	}

	soajs, err := soajsgo.NewFromConfig(context.Background(), soaConfig)
	if err != nil {
		log.Fatal(err)
	}

	router.Use(soajs.Middleware)

	router.HandleFunc("/hello", hello).Methods("GET")

	port := soaConfig.ServicePort

	go func() {
		maintenancePort := port
		if soaConfig.Maintenance.Port.Type == "maintenance" {
			maintenancePort = port + soajs.ServiceConfig.Port.MaintenanceInc
		}

		maintenanceRouter := mux.NewRouter()
		maintenanceRouter.HandleFunc("/heartbeat", Heartbeat)

		err = http.ListenAndServe(fmt.Sprintf(":%d", maintenancePort), maintenanceRouter)
		if err != nil {
			log.Fatal("maintenance services shutdown")
		}
	}()

	log.Println("starting")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
