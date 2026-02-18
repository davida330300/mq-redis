package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"mq-redis/internal/api"
	producermem "mq-redis/internal/producer/memory"
	storemem "mq-redis/internal/store/memory"
)

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	store := storemem.New()
	producer := producermem.New()
	r := api.NewRouter(store, producer)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, api.JobResponse{Status: "ok"})
	})

	server := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Printf("api listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
