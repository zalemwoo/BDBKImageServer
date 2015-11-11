package main

import (
	"bytes"
	"common"
	"encoding/json"
	"facecheck"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"time"

	"gopkg.in/fatih/pool.v2"

	"db/model"
	m "db/mongo"

	"github.com/jinzhu/configor"
)

type ServerConfig struct {
	MongoDBServer struct {
		Host     string `default:"127.0.0.1" json:"host"`
		Port     string `default:"27017" json:"port"`
		Database string `json:"database"`
	}
	ConnectionsPerServer int
	FaceRecServers       []common.ServerInfo
	ImageFilePath        string
}

var config ServerConfig
var mongo *m.Mongo

var taskPool *common.TaskPool

var processed = 0

func init() {
	err := configor.Load(&config, "./config/verify_faces.json")
	if err != nil {
		myFatal("Read config file error: ", err)
	}

	for i, _ := range config.FaceRecServers {
		if len(config.FaceRecServers[i].Url) == 0 {
			config.FaceRecServers[i].Url = config.FaceRecServers[i].Host + ":" + config.FaceRecServers[i].Port
		}
	}
}

func myFatal(message string, err error) {
	log.Fatal(message, err)
}

func shutdown() {
	fmt.Printf("SHUTDOWN.")
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
	defer shutdown()

	mongo.Session.SetCursorTimeout(0)
	mongo.Session.SetBatch(1000)

	taskPool = new(common.TaskPool)
	taskPool.Init(&common.TaskPoolConfig{
		ConnectionsPerServer: config.ConnectionsPerServer,
		Servers:              config.FaceRecServers,
	})

	go taskPool.Start()
	defer taskPool.Stop()

	image := model.ImageInfo{}
	nUnchecked := mongo.ImageFaceUncheckedOrNilCount()
	log.Printf("Unchecked images count is: %d.", nUnchecked)
	iter := mongo.ImageFaceUncheckedOrNil()
	defer iter.Close()
	for iter.Next(&image) {
		img := image
		taskPool.AddTask(func(p pool.Pool) error {
			procImage(&img, mongo, p)
			return nil
		})
	}
	_ = "breakpoint"
}

func procImage(image *model.ImageInfo, mongo *m.Mongo, p pool.Pool) {
	tcpConn, err := p.Get()
	if err != nil {
		log.Printf("Get connection from pool error. err: %v", err)
		return
	}
	defer tcpConn.Close()

	filePath := filepath.Join(config.ImageFilePath, image.File_path)
	message := facecheck.BuildMessage(filePath)
	if message == nil {
		log.Printf("Build message Error.")
		return
	}

	messageLen := len(message)
	totalWrite := 0
	for {
		tcpConn.SetWriteDeadline(time.Now().Add(60 * time.Second))
		n, err := tcpConn.Write(message[totalWrite:])
		if err != nil {
			log.Printf("Write message error. err: %v", err)
			return
		}
		totalWrite += n
		if totalWrite >= messageLen {
			break
		}
	}

	dataBuffer := bytes.Buffer{}
	b := make([]byte, 256)
	totalLen := int32(18)
	messageBody := ""
	isComplete := false
	for {
		tcpConn.SetReadDeadline(time.Now().Add(60 * time.Second))
		n, err := tcpConn.Read(b)
		if err != nil {
			if err != io.EOF {
				log.Printf("Read message from FaceRec server Error. err: = %v", err)
			}
			return
		}
		dataBuffer.Write(b[:n])
		if int32(dataBuffer.Len()) >= totalLen {
			totalLen, messageBody, isComplete = facecheck.ParseResult(dataBuffer.Bytes())
			if isComplete {
				break
			}
		}
	}
	image.Has_face = (len(messageBody) != 0)
	image.Face_info = messageBody
	image.Face_checked = true

	go func(mongo *m.Mongo, image *model.ImageInfo) {
		mongo.ImageUpdateHasFace(image)
	}(mongo, image)

	processed++
	if image.Has_face == true {
		log.Printf("%d: FOUND FACE: %s", processed, filePath)
	} else {
		log.Printf("%d: NO FACE: %s", processed, filePath)
	}
}
