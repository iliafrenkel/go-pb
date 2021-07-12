package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iliafrenkel/go-pb/src/api"
)

var webServer *http.Server

func StartWebServer() error {
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

	router.LoadHTMLGlob(filepath.Join("..", "src", "web", "templates", "*.html"))

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

	router.GET("/p/:id", func(c *gin.Context) {
		var p api.Paste
		id := c.Param("id")
		resp, err := http.Get("http://localhost:8080/paste/" + id)

		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "unexpected api error")
			return
		}
		// Check API response status
		if resp.StatusCode != http.StatusOK {
			c.String(resp.StatusCode, "api: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
			return
		}
		// Get the paste from the body
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "unexpected api error")
			return
		}
		// Try to parse JSON into api.Paste
		if err := json.Unmarshal(b, &p); err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "failed to parse api response")
			return
		}
		// Send HTML
		c.HTML(
			http.StatusOK,
			"view.html",
			gin.H{
				"Title":    p.Title,
				"Body":     p.Body,
				"Language": p.Syntax,
				"URL":      p.URL(),
				"Server":   "http://localhost:8000", //TODO: this has to come from somewhere
			},
		)
	})

	router.POST("/p/", func(c *gin.Context) {
		var p api.Paste
		var data api.Paste
		// Get the paste title and body from the form
		if b, ok := c.GetPostForm("body"); !ok || len(b) == 0 {
			c.String(http.StatusBadRequest, "body cannot be empty")
			return
		}
		p.Body = c.PostForm("body")
		p.Title = c.DefaultPostForm("title", "untitled")
		p.DeleteAfterRead, _ = strconv.ParseBool(c.PostForm("delete_after_read"))
		p.Syntax = c.DefaultPostForm("syntax", "none")

		// Try to create a new paste by calling the API
		paste, _ := json.Marshal(p)
		resp, err := http.Post("http://localhost:8080/paste", "application/json", bytes.NewBuffer(paste))

		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "unexpected api error")
			return
		}

		// Check API response status
		if resp.StatusCode != http.StatusCreated {
			c.String(resp.StatusCode, "api: %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
			return
		}

		// Get API response body
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "unexpected error parsing api response")
			return
		}

		if err := json.Unmarshal(b, &data); err != nil {
			log.Println(err)
			c.String(http.StatusInternalServerError, "failed to parse api response")
			return
		}

		c.HTML(
			http.StatusOK,
			"view.html",
			gin.H{
				"Title":    data.Title,
				"Body":     data.Body,
				"Language": data.Syntax,
				"URL":      resp.Header.Get("Location"),
				"Server":   "http://localhost:8000", //TODO: this has to come from somewhere
			},
		)
	})

	router.Static("/assets", "../src/web/assets")

	addr := "127.0.0.1:8000"
	webServer := &http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	log.Println("Web server listening on ", addr)

	return webServer.ListenAndServe()
}

func StopWebServer(ctx context.Context) error {
	if webServer != nil {
		return webServer.Shutdown(ctx)
	}

	return nil
}
