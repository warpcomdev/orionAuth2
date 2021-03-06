package main

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
)

func main() {

	api := NewApi(
		"/login", "/accept", "/consent", "/error",
		nil,
		1*time.Hour,
		&HydraClient{
			URL: "http://localhost:8080",
		},
		&OrionClient{},
		&JWT{
			Issuer:        "testIssuer",
			SigningMethod: jwt.SigningMethodHS512,
			Keyfunc: func(*jwt.Token) (interface{}, error) {
				return []byte("12345678901234567890123456789012"), nil
			},
		})

	csrfKey := []byte("12345678901234567890123456789012")
	router := mux.NewRouter()
	apiRoute := router.Path("/api").Subrouter()
	csrfprot := csrf.Protect(
		csrfKey,
		csrf.SameSite(csrf.SameSiteStrictMode),
		csrf.ErrorHandler(http.HandlerFunc(api.csrfErrorHandler)))
	apiRoute.Use(func(handler http.Handler) http.Handler {
		return csrfprot(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("X-CSRF-Token", string(csrf.Token(r)))
			handler.ServeHTTP(w, r)
		}))
	})

	apiRoute.HandleFunc("/auth", api.PostLogin).Methods("POST")
	apiRoute.HandleFunc("/auth", api.GetLogin).Methods("GET")
	apiRoute.HandleFunc("/consent", api.PostConsent).Methods("POST")
	apiRoute.HandleFunc("/consent", api.GetConsent).Methods("GET")

	http.ListenAndServe(":8080", router)
}
