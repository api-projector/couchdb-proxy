package server

import (
	"couchdb-proxy/internal/pg"
	"couchdb-proxy/internal/proxy"
	"errors"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"strings"
)

const (
	serverPort = ":8080"
	authHeader = "Authorization"
	authPrefix = "Bearer "
)

var couchDbProxy *proxy.CouchDbProxy
var pgPool *pgxpool.Pool

func Run() {
	http.HandleFunc("/", handler)

	log.Printf("starting http server on port %s ...", serverPort)
	log.Fatal(http.ListenAndServe(serverPort, nil))
}

func init() {
	viper.SetEnvPrefix("proxy")

	couchDbProxy = proxy.NewCouchDbProxy()
	pgPool = pg.GetConnectionPool()
}

func handler(writer http.ResponseWriter, request *http.Request) {
	authToken := extractAuthToken(request)
	database, err := extractDatabase(request)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	err = couchDbProxy.ProxyRequest(pgPool, authToken, database, writer, request)
	if err != nil {
		switch err.(type) {
		case *proxy.ForbiddenError:
			writer.WriteHeader(http.StatusForbidden)
		default:
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
	}
}

func extractAuthToken(request *http.Request) string {
	auth := request.Header.Get(authHeader)

	return strings.TrimPrefix(auth, authPrefix)
}

func extractDatabase(request *http.Request) (string, error) {
	parts := strings.Split(request.RequestURI, "/")
	if len(parts) <= 1 {
		return "", errors.New("can't determine database")
	}

	return parts[1], nil
}
