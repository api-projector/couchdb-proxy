package proxy

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	paramCouchDbUrl   = "couchdb_url"
	paramCouchDbUser  = "couchdb_user"
	paramCouchDbRoles = "couchdb_roles"
)

var couchDbParams = []string{paramCouchDbUrl, paramCouchDbUser, paramCouchDbRoles}

type couchDbConfig struct {
	Url   *url.URL
	User  string
	Roles string
}

type CouchDbProxy struct {
	config       *couchDbConfig
	reverseProxy *httputil.ReverseProxy
}

type ForbiddenError struct{}

func (err *ForbiddenError) Error() string { return "Forbidden" }

func (proxy *CouchDbProxy) ProxyRequest(pool *pgxpool.Pool, authToken string, database string, writer http.ResponseWriter, request *http.Request) (err error) {
	allowed, err := isAccessAllowed(pool, database, authToken)
	if err != nil {
		return
	}

	if !allowed {
		return &ForbiddenError{}
	}

	request.Header["X-Auth-CouchDB-Roles"] = []string{proxy.config.Roles}
	request.Header["X-Auth-CouchDB-UserName"] = []string{proxy.config.User}

	log.Printf("proxy request: auth=%s, db=%s", authToken, database)

	proxy.reverseProxy.ServeHTTP(writer, request)
	return
}

func NewCouchDbProxy() *CouchDbProxy {
	config := readCouchDbConfig()

	log.Printf("create proxy proxy: url=%s, user=%s", config.Url, config.User)

	reverseProxy := httputil.NewSingleHostReverseProxy(config.Url)
	reverseProxy.FlushInterval = 500 * time.Millisecond

	return &CouchDbProxy{
		config:       config,
		reverseProxy: reverseProxy,
	}
}

func readCouchDbConfig() *couchDbConfig {
	for _, env := range couchDbParams {
		err := viper.BindEnv(env)
		if err != nil {
			panic(err)
		}
	}

	viper.SetDefault(paramCouchDbUrl, "http://couchdb:5984")
	viper.SetDefault(paramCouchDbUser, "admin")
	viper.SetDefault(paramCouchDbRoles, "_admin")

	couchdbUrl, err := url.Parse(viper.GetString(paramCouchDbUrl))
	if err != nil {
		panic(err)
	}

	return &couchDbConfig{
		Url:   couchdbUrl,
		User:  viper.GetString(paramCouchDbUser),
		Roles: viper.GetString(paramCouchDbRoles),
	}
}
