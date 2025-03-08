package main

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed *.tmpl
var templateFS embed.FS
var tmpl = template.Must(template.ParseFS(templateFS, "*.tmpl"))

//go:embed style.css
var stylesheet string

func StartServer(addr string) error {
	http.HandleFunc("/c/", channelHandler)
	http.HandleFunc("/user", userHandler)
	http.HandleFunc("/", rootHandler)
	return http.ListenAndServe(addr, nil)
}
