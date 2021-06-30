package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/imrushi/oauth2-openid-hydra/pkg/tracer"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/urfave/negroni"
	goauth2 "golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var Endpoint = "http://localhost:4444/oauth2/token"

var OAuthConf = &clientcredentials.Config{
	ClientID:     os.Getenv("CLIENT_ID"),
	ClientSecret: os.Getenv("CLIENT_SECRET"),
	Scopes:       []string{"read", "write"},
	TokenURL:     Endpoint,
}

func main() {
	// Prepare Opentracing
	var (
		tracerServiceName     = "ClientCred"
		tracerURL             = "localhost:6831"
		tracerService, closer = tracer.New(true, tracerServiceName, tracerURL, 1)
	)

	defer func() {
		if closer == nil {
			_, _ = fmt.Fprintf(os.Stderr, "tracer closer is nil\n")
			return
		}

		if err := closer.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "closing tracer error: %s\n", err.Error())
			return
		}
	}()

	// set global tracer of this application
	opentracing.SetGlobalTracer(tracerService)

	r := mux.NewRouter()

	r.HandleFunc("/client_cred", ClientEndpoint(*OAuthConf)).Methods("GET")

	// Set up a request logger, useful for debugging
	n := negroni.New()
	n.Use(negronilogrus.NewMiddleware())
	n.UseHandler(r)

	if err := http.ListenAndServe(":8001", r); err != nil {
		log.Fatal(err)
	}
}

func ClientEndpoint(c clientcredentials.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := getToken(c)
		if err != nil {
			str := fmt.Sprint("Received an error: ", err.Error())
			http.Error(w, str, http.StatusUnprocessableEntity)
		}
		renderTemplate(w, "client_cred.html", map[string]interface{}{
			"Token":  token,
			"Scope":  token.Extra("scope"),
			"Expiry": token.Extra("expires_in"),
		})
	}
}

var getToken = func(conf clientcredentials.Config) (*goauth2.Token, error) {
	conf.AuthStyle = goauth2.AuthStyleAutoDetect
	return conf.Token(context.Background())
}

// renderTemplate is a convenience helper for rendering templates.
func renderTemplate(w http.ResponseWriter, id string, d interface{}) bool {
	if t, err := template.New(id).ParseFiles("../frontend/templates/" + id); err != nil {
		http.Error(w, errors.Wrap(err, "Could not render template").Error(), http.StatusInternalServerError)
		return false
	} else if err := t.Execute(w, d); err != nil {
		http.Error(w, errors.Wrap(err, "Could not render template").Error(), http.StatusInternalServerError)
		return false
	}
	return true
}
