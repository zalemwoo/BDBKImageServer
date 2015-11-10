package main

import (
	"bufio"
	"bytes"
	"common"
	"encoding/json"
	"facecheck"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"utils"

	"gopkg.in/fatih/pool.v2"

	"db/model"
	m "db/mongo"

	"github.com/codegangsta/cli"
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

var wg sync.WaitGroup

func init() {
	err := configor.Load(&config, "./config/check_face.json")
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
	fmt.Println("SHUTDOWN.")
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
	mongo.Session.SetBatch(10000)

	taskPool = new(common.TaskPool)
	taskPool.Init(&common.TaskPoolConfig{
		ConnectionsPerServer: config.ConnectionsPerServer,
		Servers:              config.FaceRecServers,
	})

	go taskPool.Start()
	defer taskPool.Stop()

	app := cli.NewApp()
	app.Name = "check_face"
	app.Usage = "check face in image"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Value: "",
			Usage: "name of image file.",
		},
	}

	app.Action = func(c *cli.Context) {
		name := c.String("name")
		if len(name) == 0 {
			println("name are required.")
			cli.ShowAppHelp(c)
			return
		}
		iter := mongo.ImageByName(name)
		image := model.ImageInfo{}
		for iter.Next(&image) {
			wg.Add(1)
			img := image
			log.Printf("Found image. image: %v", img)
			taskPool.AddTask(func(p pool.Pool) error {
				procImage(&img, mongo, p)
				wg.Done()
				return nil
			})
			wg.Wait()
		}
		iter.Close()
	}

	app.Run(os.Args)
}

func procImage(img *model.ImageInfo, mongo *m.Mongo, p pool.Pool) {
	image_json, _ := json.Marshal(img)
	log.Printf("Processing image. image: %s", image_json)
	tcpConn, err := p.Get()
	if err != nil {
		log.Printf("Get connection from pool error. err: %v", err)
		return
	}
	defer tcpConn.Close()

	filePath := filepath.Join(config.ImageFilePath, img.File_path)
	log.Printf("Processing. path: %s", filePath)
	message := facecheck.BuildMessage(filePath)
	if message == nil {
		log.Printf("Build message Error.")
		return
	}

	writer := bufio.NewWriter(tcpConn)
	n, err := writer.Write(message)
	if err != nil {
		log.Printf("Write message to FaceRec server Error. err: = %v", err)
		return
	}
	if n != len(message) {
		log.Printf("Write message to FaceRec size Error. expect: %d, wrote: %d", len(message), n)
		return
	}
	err = writer.Flush()
	if err != nil {
		log.Printf("Flush message to FaceRec server Error. err: = %v", err)
		return
	}

	dataBuffer := bytes.Buffer{}
	b := make([]byte, 256)
	totalLen := int32(18)
	messageBody := ""
	isComplete := false
	for {
		n, err := tcpConn.Read(b)
		if err != nil {
			if err != io.EOF {
				log.Printf("Read message from FaceRec server Error. err: = %v", err)
			}
			return
		}
		dataBuffer.Write(b[:n])
		log.Printf("Read message from FaceRec server. n: %d, buffer: %x", n, dataBuffer.Bytes())
		if int32(dataBuffer.Len()) >= totalLen {
			totalLen, messageBody, isComplete = facecheck.ParseResult(dataBuffer.Bytes())
			log.Printf(`Read message from FaceRec server. totalLen: %d, messageBody: "%s", isComplete: %v`, totalLen, messageBody, isComplete)
			if isComplete {
				break
			}
		}
	}
	img.Has_face = (len(messageBody) != 0)
	img.Face_info = messageBody
	img.Face_checked = true

	if img.Has_face == false {
		log.Printf("NO FACE: %v", img)
		return
	} else {
		log.Printf("FOUND FACE: %v", img)
	}

	file_in, err := os.Open(filePath)
	if err != nil {
		log.Fatal("File open error. err: ", err)
		os.Exit(255)
	}
	defer file_in.Close()

	image_in, err := jpeg.Decode(file_in)
	if err != nil {
		log.Fatal("Image decode error. err: ", err)
		os.Exit(255)
	}
	file_out, err := os.Create("Original.jpg")
	if err != nil {
		log.Fatal("File create error. err: ", err)
		os.Exit(255)
	}
	defer file_out.Close()
	jpeg.Encode(file_out, image_in, &jpeg.Options{100})

	faceInfo, err := facecheck.ParseFaceInfo(img.Face_info)
	if err != nil {
		fmt.Printf("Parse Rect error. err: %v, rect: %s\n", err, img.Face_info)
	}
	faceInfo_json, _ := json.Marshal(faceInfo)
	fmt.Printf("Face info is: %s\n", faceInfo_json)

	writeImageFunc := func(r *facecheck.Rect, suffix string, i int) {
		fmt.Printf("%s rect is: %v\n", suffix, r)
		rect := image.Rect(r.Left, r.Top, r.Right, r.Bottom)
		file_out, err := os.Create("Out_" + suffix + strconv.Itoa(i) + ".jpg")
		if err != nil {
			log.Fatal("File create error. err: ", err)
			os.Exit(255)
		}
		defer file_out.Close()
		croped := utils.CropImage(image_in, rect)
		jpeg.Encode(file_out, croped, &jpeg.Options{100})
	}

	for i, v := range faceInfo.Faces {
		writeImageFunc(&v, "Face", i)
	}
	for i, v := range faceInfo.Heads {
		writeImageFunc(&v, "Head", i)
	}
}
