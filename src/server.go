package main

import (
	"encoding/json"
	"log"

	"api"
	m "db/mongo"

	"github.com/jinzhu/configor"
)

type ServerConfig struct {
	Server struct {
		Port string `default:"12346" json:"port"`
	}
	MongoDBServer struct {
		Host     string `default:"127.0.0.1" json:"host"`
		Port     string `default:"27017" json:"port"`
		Database string `json:"database"`
	}
	FileServer struct {
		URL string `default:"http://127.0.0.1/" json:"url"`
	}
}

var config ServerConfig
var mongo *m.Mongo

func init() {
	err := configor.Load(&config, "./config/server_config.json")
	if err != nil {
		myFatal("Read config file error: ", err)
	}
}

func myFatal(message string, err error) {
	log.Fatal(message, err)
}

func shutdown() {
	if mongo != nil && mongo.Connected {
		mongo.Close()
		mongo = nil
	}
}

func main() {
	config_json, _ := json.Marshal(config)
	log.Printf("Server config: %s \n", config_json)

	mongoConfig := m.MongoConfig{
		Host:     config.MongoDBServer.Host,
		Port:     config.MongoDBServer.Port,
		Database: config.MongoDBServer.Database,
	}

	mongo, err := m.NewMongo(&mongoConfig)

	if err != nil {
		myFatal("Connect to mongodb server Error. err: = %v", err)
	} else {
		log.Printf("Connect to mongodb [%v]: %s, db: %s.", mongo.Connected, mongo.ServerUrl, mongo.DatabaseName)
	}

	restConfig := api.RESTfulConfig{
		Port:          config.Server.Port,
		FileServerURL: config.FileServer.URL,
		Mongo:         mongo,
	}

	restService := api.CreateService(&restConfig)
	err = restService.Serve()
	if err != nil {
		myFatal("Server start error!!! err: ", err)
	}

	defer shutdown()

}
