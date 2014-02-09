package main

import (
	"io"
	"fmt"
	"bytes"
	"log"
	"flag"
	"net/http"
	"encoding/json"
	"strings"
)

// API Docs
// http://apidocs.mailchimp.com/api/2.0/lists/subscribe.php

type MCEmail struct {
	Email string				`json:"email"`
}

type MCRequest struct {
	EmailInfo MCEmail `json:"email"`
	Id string					`json:"id"`
	ApiKey string			`json:"apikey"`
}


// Example responses
// {status: "error", code: -100, name: "ValidationError", error: "An email address must contain a single @"}
// {status: "error", code: -100, name: "ValidationError", error: "This email address looks fake or invalid. Please enter a real email address"}
// {email: "good@email.com", euid: "eid", leid: "leid"}

var (
	mc_api_key *string = flag.String("mc-api-key", "", "Mailchimp API key")
	mc_list_id *string = flag.String("mc-list-id", "", "Mailchimp list id")
	listen		 *string = flag.String("l", "127.0.0.1:3001", "Address to listen on")
)

func mc_subscribe(email string) (*http.Response, error) {

	mcr := MCRequest{
		EmailInfo: MCEmail{ Email: email },
		Id: *mc_list_id,
		ApiKey: *mc_api_key,
	}

	parts := strings.Split(*mc_api_key, "-")
	dc := parts[1]
	path := "lists/subscribe.json"
	api_url :=  fmt.Sprintf("https://%s.api.mailchimp.com/2.0/%s", dc, path)

	data, err := json.Marshal(mcr)

	if err != nil {
		return nil, err
	}

	r, err := http.Post(api_url, "application/json", bytes.NewBuffer(data))

	return r, err
}

func subscribe(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-type", "application/json")

	email := req.FormValue("email")

	if "" == email {
		z := []byte(`{"status":"error","name":"ValidationError","error":"Please enter your email"}`)
		io.Copy(w, bytes.NewBuffer(z))
		return
	}

	r, err:= mc_subscribe(email)

	if err == nil {
		log.Println("subscribing", email)
	} else {
		log.Printf("Error subscribing %s: %s", email, err.Error())
	}

	// z := []byte(`{"status": "error", "code": -100, "name": "ValidationError", "error": "An email address must contain a single @"}`)
	// z := []byte(`{"email": "good@email.com", "euid": "eid", "leid": "leid"}`)
	// <-time.After(time.Second * 2)
	// _, err = io.Copy(w, bytes.NewBuffer(z))
	_, err = io.Copy(w, r.Body)
}

func subscribeMux(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Header().Set("Location", "/")
		w.WriteHeader(302)
	case "POST":
		subscribe(w,r)
	}
}

func main() {
	flag.Parse()

	staticServer := http.FileServer(http.Dir("./public/"))

	http.HandleFunc("/",func(w http.ResponseWriter, r *http.Request){ http.ServeFile(w,r, "./public/index.html") })
	http.Handle("/js/", staticServer)
	http.Handle("/css/",staticServer)

	http.HandleFunc("/subscriptions/", subscribeMux)

	log.Printf("Serving at %s", *listen)

	log.Fatal(http.ListenAndServe(*listen, nil))
}

