package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	m "db/mongo"

	"github.com/gorilla/mux"
)

type RESTfulConfig struct {
	Port          string
	FileServerURL string
	Mongo         *m.Mongo
}

type RESTfulService struct {
	Config RESTfulConfig
	mongo  *m.Mongo
	router *mux.Router
	paths  []string
}

var service *RESTfulService

func (self *RESTfulService) RegisterHandler(path string, f func(http.ResponseWriter, *http.Request)) {
	self.paths = append(service.paths, path)
	self.router.Path(path).HandlerFunc(f)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	routes_json, _ := json.Marshal(service.paths)
	fmt.Fprintf(w, "All Routes: %s", routes_json)
}

func (self *RESTfulService) Serve() error {
	log.Printf("Server start serve at %s", self.Config.Port)
	http.Handle("/", self.router)
	return http.ListenAndServe(":"+self.Config.Port, nil)
}

func CreateService(config *RESTfulConfig) *RESTfulService {
	service = new(RESTfulService)
	service.Config = *config
	service.mongo = config.Mongo
	service.router = mux.NewRouter()
	service.paths = make([]string, 0, 16)

	service.RegisterHandler("/", rootHandler)

	return service
}
