package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

func main() {

	// Create http endpoints for external consumption

	r := mux.NewRouter() // gorilla/mux router

	// Simple status check
	r.Handle("/status", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API is up and running"))
	})).Methods("GET")

	// Sign in
	r.HandleFunc("/signin", signin) // creates jwt cookie for valid user, request body should have email/password in JSON format

	// Validate token
	r.Handle("/validatetoken", AuthMiddleware(validatetoken)).Methods("GET")

	// Secured handlers ahould use AuthMiddleware to check token
	// Get all users
	r.Handle("/users", AuthMiddleware(getUsers)).Methods("GET")

	// Get specific user
	r.Handle("/user/{email}", AuthMiddleware(getUser)).Methods("GET")

	// Add (register) a user
	r.Handle("/user", putUser).Methods("PUT")

	// Update specific user
	r.Handle("/user", AuthMiddleware(postUser)).Methods("POST")

	// Delete a specific user
	r.Handle("/user/{email}", AuthMiddleware(deleteUser)).Methods("DELETE")

	// Listen with the Logging handler from gorilla
	http.ListenAndServe(":3000", handlers.LoggingHandler(os.Stdout, r))
}

func signin(w http.ResponseWriter, r *http.Request) {

	// Get user from database, if found return jwt cookie

	// Decode JSON body into credentials (email and password)
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := GetValidClaims(creds)

	if claims.IsValid {
		expirationTime := time.Now().Add(5 * time.Minute) // Declare expiration time of five minutes

		// Create token with these claims and set cookie
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(MySigningKey)
		if err != nil {
			log.Printf("SignedString error: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Set a cookie with the new token
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
		})
	} else {
		log.Print("SIGNIN NOT VALID")
		w.WriteHeader(http.StatusUnauthorized)
	}

	log.Printf("SIGNIN for %s", creds.Email)
}

var validatetoken = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	// AuthMiddleware validates for us

	w.WriteHeader(http.StatusOK)
})

var getUsers = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	// get all users from database and return as JSON array

	// Convert key-value store to array of User
	users := make([]User, memCache.ItemCount())
	n := 0

	for _, v := range memCache.Items() {
		users[n] = v.Object.(User)
		n++
	}
	payload, _ := json.Marshal(users)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(payload))
	log.Print("GET USERS")
})

var putUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	// Create new non-admin user in database, NB not authenticated

	// Get JSON body and decode into user
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check email is unique
	_, ok := memCache.Items()[user.Email]
	if ok {
		w.WriteHeader(http.StatusConflict)
		return
	}

	memCache.Set(user.Email, user, cache.DefaultExpiration)
	log.Printf("PUT USER: %s", user.Email)
})

var getUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	// Get specific user

	vars := mux.Vars(r)
	email := vars["email"]
	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u, ok := memCache.Items()[email]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user := u.Object.(User)
	payload, _ := json.Marshal(user)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(payload))
	log.Printf("GET USER: %s", user.Email)
})

var postUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	// Update specific user in database, NB admin only

	// Get JSON body and decode into user
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check exists
	_, ok := memCache.Items()[user.Email]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	memCache.Set(user.Email, user, cache.DefaultExpiration)
	log.Printf("POST USER: %s", user.Email)

})

var deleteUser = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	// Delete specific user in database, NB admin only

	vars := mux.Vars(r)
	email := vars["email"]
	if email == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, ok := memCache.Items()[email]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	memCache.Delete(email)

	log.Printf("DELETED USER: %s", email)

})
