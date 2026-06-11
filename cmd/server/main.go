package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	goadminer "github.com/Asendar1/GoAdminer"
	"github.com/Asendar1/GoAdminer/internal/handler"
	"github.com/Asendar1/GoAdminer/internal/server"
	"github.com/Asendar1/GoAdminer/internal/session"
)

func main() {
	port := flag.String("port", "8080", "HTTP port")
	dev := flag.Bool("dev", false, "Serve frontend from disk (web/)")
	flag.Parse()

	sessions := session.NewStore()
	h := handler.New(sessions)

	var srv http.Handler
	if *dev {
		srv = server.New(h, nil, true)
		log.Println("Dev mode: serving frontend from web/")
	} else {
		srv = server.New(h, &goadminer.WebFS, false)
	}

	httpSrv := &http.Server{
		Addr:         ":" + *port,
		Handler:      srv,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("GoAdminer listening on %s", ":"+*port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exited")
}
