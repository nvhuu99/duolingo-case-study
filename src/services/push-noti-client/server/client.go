package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

func home(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
	file, err := os.Open("../public/client.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	
	if _, err := io.Copy(w, file); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func saveToken(w http.ResponseWriter, r *http.Request) {
    var req struct {
		Token string `json:"token"`
	}
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err)
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

	file, err := os.OpenFile("../storage/registrationTokens", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	_, err = file.WriteString(req.Token + "\n")
	if err != nil {
		log.Println(err)
		http.Error(w, "Failed to write to file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	w.Write([]byte("Token saved"))
}

func main() {
    http.HandleFunc("/home", home)
    http.HandleFunc("/save-token", saveToken)
    http.Handle("/", http.FileServer(http.Dir("../public")))
    http.ListenAndServe(":8004", nil)
}