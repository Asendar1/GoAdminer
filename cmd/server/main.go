package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"goadminer"
	"goadminer/internal/handler"
	"goadminer/internal/server"
	"goadminer/internal/session"
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

	addr := ":" + *port
	log.Printf("GoAdminer listening on %s", addr)

	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
