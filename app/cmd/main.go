package main

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	// // Load templates
	// pattern := filepath.Join("..", "templates", "*.html")
	// templates := template.Must(template.ParseGlob(pattern))

	// // Define static assets location
	// r := mux.NewRouter()
	// staticFilesDirectory := http.Dir("../assets/")
	// staticFileHandler := http.StripPrefix("/assets/", http.FileServer(staticFilesDirectory))
	// r.PathPrefix("/assets/").Handler(staticFileHandler).Methods("GET")

	// // Define routes
	// r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	if err := templates.ExecuteTemplate(w, "index", nil); err != nil {
	// 		log.Fatalf("Failed to execute index template: %s", err)
	// 	}

	// }).Methods("GET")

	// log.Fatal(http.ListenAndServe(":8080", r))

	router := gin.Default()

	router.LoadHTMLGlob(filepath.Join("..", "templates", "*.html"))

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.HTML(
			http.StatusOK,
			"index.html",
			gin.H{
				"title": " Go PB - Home",
			},
		)
	})

	router.GET("/u/login", func(c *gin.Context) {
		c.HTML(
			http.StatusOK,
			"login.html",
			gin.H{
				"title": " Go PB - Login",
			},
		)
	})

	router.GET("/u/register", func(c *gin.Context) {
		c.HTML(
			http.StatusOK,
			"register.html",
			gin.H{
				"title": " Go PB - Register",
			},
		)
	})

	router.Run("127.0.0.1:8080")
}
