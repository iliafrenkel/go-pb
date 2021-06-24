package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	staticFilesDirectory := http.Dir("../assets/")
	staticFileHandler := http.StripPrefix("/assets/", http.FileServer(staticFilesDirectory))
	r.PathPrefix("/assets/").Handler(staticFileHandler).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", r))
}
