package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/imrushi/oauth2-openid-hydra/cmd/authgf/handler"
	"github.com/imrushi/oauth2-openid-hydra/cmd/authgf/ldapclient"
	"github.com/imrushi/oauth2-openid-hydra/pkg/tracer"

	"github.com/gorilla/mux"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/opentracing/opentracing-go"
	hydra "github.com/ory/hydra-client-go/client"
	"github.com/urfave/negroni"
)

var (
	adminURL, _ = url.Parse("http://localhost:4445")
	hydraClient = hydra.NewHTTPClientWithConfig(nil,
		&hydra.TransportConfig{
			Schemes:  []string{adminURL.Scheme},
			Host:     adminURL.Host,
			BasePath: adminURL.Path,
		},
	)
)

type Config struct {
	LDAP ldapclient.Config
}

// var userInfo = []repouser.UserInfo{
// 	{
// 		ID:       1,
// 		Email:    "user@example.com",
// 		Password: "password",
// 	},
// 	{
// 		ID:       2,
// 		Email:    "user2@example.com",
// 		Password: "password",
// 	},
// }

func main() {
	var cnf Config
	//prepare Opentracing
	var (
		tracerServiceName     = "AuthGF"
		tracerURL             = "localhost:6831"
		tracerService, closer = tracer.New(true, tracerServiceName, tracerURL, 1)
	)

	defer func() {
		if closer == nil {
			_, _ = fmt.Fprintf(os.Stderr, "tracer closer is nil\n")
		}

		if err := closer.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stdout, "closing tracer error: %s\n", err.Error())
		}
	}()
	cnf.LDAP.Endpoints = append(cnf.LDAP.Endpoints, "<ad serverIP>:636")
	cnf.LDAP.BaseDN = "dc=example,dc=com"
	cnf.LDAP.RoleBaseDN = "ou=Users,dc=example,dc=com"
	cnf.LDAP.BindDN = "abc@example.com"
	cnf.LDAP.BindPass = "p@ssw0rd12#"
	cnf.LDAP.IsTLS = true
	cnf.LDAP.AttrClaims = map[string]string{"name": "name", "sn": "family_name", "givenName": "given_name", "mail": "email"}

	fmt.Print(cnf.LDAP.Endpoints)
	//set global tracer of this application
	opentracing.SetGlobalTracer(tracerService)

	ldap := ldapclient.New(cnf.LDAP)
	controller := handler.Handler{
		HydraAdmin: hydraClient.Admin,
		// UserRepo:   repouser.NewMemory(userInfo),
		Um: ldap,
	}

	r := mux.NewRouter()
	// Set up a request logger, useful for debugging
	n := negroni.New()
	n.Use(negronilogrus.NewMiddleware())
	n.UseHandler(r)

	// r.Use(jaegertracing.TraceWithConfig(jaegertracing.TraceConfig{
	// 	Tracer: tracerService,
	// }))

	r.HandleFunc("/authentication/login", controller.GetLogin).Methods("GET")
	r.HandleFunc("/authentication/login", controller.PostLogin).Methods("POST")
	r.HandleFunc("/authentication/consent", controller.GetConsent).Methods("GET")
	r.HandleFunc("/authentication/consent", controller.PostConsent).Methods("POST")

	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatal(err)
	}
}
