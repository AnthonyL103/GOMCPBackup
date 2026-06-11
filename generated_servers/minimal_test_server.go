package main

import (
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/execute/", handleHealth)
	http.HandleFunc("/execute/test", handleTest)

	log.Printf("Starting minimal_test on port 9001")
	log.Fatal(http.ListenAndServe(":9001", nil))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func handleTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

		result := map[string]string{"status": "ok"}
		data, _ := json.Marshal(result)
		w.Write(data)
}

