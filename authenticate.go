package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"

	// "database/sql"

	// _ "github.com/go-sql-driver/mysql"

	"github.com/spf13/viper"

	"github.com/dgrijalva/jwt-go"
)

// MySigningKey is in config
var MySigningKey = []byte("")

// Credentials from the initial login request
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Claims go in the token
type Claims struct {
	Email   string `json:"email"`
	IsAdmin bool   `json:"isAdmin"`
	IsValid bool   `json:"isValid"`
	jwt.StandardClaims
}

var memCache *cache.Cache

func init() {
	// Get config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}
	viper.SetDefault("AdminEmail", "admin@ecn.com")
	viper.SetDefault("AdminPassword", "pwd")
	viper.SetDefault("MySigningKey", "SECRET_KEY")

	MySigningKey = []byte(viper.GetString("MySigningKey"))
	// TODO create admin user in database
	// db, err := sql.Open("mysql", "user:password@/dbname")
	// if err != nil {
	// 	panic(fmt.Errorf("fatal error opeing database: %s", err))
	// }
	// db.Ping()

	// Use an in-mem cache for simplicity and add our admin user
	adminUser := User{
		Email:    viper.GetString("AdminEmail"),
		LastName: "admin",
		Password: viper.GetString("AdminPassword"),
	}

	memCache = cache.New(24*time.Hour, 1*time.Hour)
	memCache.Set(adminUser.Email, adminUser, cache.DefaultExpiration)

	// Add dummy users for testing
	for _, user := range users {
		//log.Print(user.Email)
		memCache.Set(user.Email, user, cache.DefaultExpiration)
	}

	log.Print("memCache intialized")
	log.Print(memCache.Items())
}

// GetValidClaims creates claims for a vaild user
func GetValidClaims(creds Credentials) *Claims {

	claims := &Claims{IsValid: false}

	// Look for user in database and build claims for token
	if x, found := memCache.Get(creds.Email); found {
		user := x.(User)

		if creds.Password == user.Password {
			// Is a valid user
			claims.IsValid = true
			claims.Email = user.Email

			// Check if user is a configured admin
			if user.Email == viper.GetString("AdminEmail") {
				log.Print("ADMIN USER")
				claims.IsAdmin = true
			}
		}
	}
	return claims
}

// AuthMiddleware handles authenticating the user
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print("executing auth middleware")

		_, err := authenticate(w, r)
		if err != nil {
			return
		}

		next.ServeHTTP(w, r)

	})
}

func authenticate(w http.ResponseWriter, r *http.Request) (*Claims, error) {
	// Get the claims token from the request's cookies
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return nil, err
		}
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	// Get the JWT string
	tknStr := c.Value
	//fmt.Print(tknStr)

	// New claims instance
	claims := &Claims{}

	// Parse and store string
	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return MySigningKey, nil
	})
	if err != nil {
		log.Printf("authenticate error: %s", err.Error())
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return nil, err
		}
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return nil, err
	}

	log.Print("AUTHENTICATED")
	return claims, nil
}
