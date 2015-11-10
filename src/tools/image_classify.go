package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"utils"

	"db/model"
	m "db/mongo"

	"github.com/jinzhu/configor"
)

const (
	JOBS              = 20
	FILES_PER_DIR     = 10000
	HAS_FACE_FILE_DIR = "face"
	NO_FACE_FILE_DIR  = "no_face"
)

type ServerConfig struct {
	MongoDBServer struct {
		Host     string `default:"127.0.0.1" json:"host"`
		Port     string `default:"27017" json:"port"`
		Database string `json:"database"`
	}
	ImageFilePath string `default:"127.0.0.1" json:"image_file_path"`
	TargetDirPath string `default:"./tmp" json:"target_dir_path"`
}

var config ServerConfig
var mongo *m.Mongo
var wg sync.WaitGroup

var has_face_image_cnt = 0
var no_face_image_cnt = 0

func init() {
	err := configor.Load(&config, "./config/image_classify.json")
	if err != nil {
		myFatal("Read config file error: ", err)
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
	mongo.Session.SetBatch(10000)

	iter := mongo.ImageFaceChecked(true)
	has_image := true
	for {
		if has_image == false {
			break
		}
		wg.Add(JOBS)
		for i := 0; i < JOBS; i++ {
			image := model.ImageInfo{}
			if iter.Next(&image) {
				var curr_count *int
				curr_base_dir := ""
				if image.Has_face == true {
					log.Printf("FACE FIlE: %s", image.File_path)
					curr_count = &has_face_image_cnt
					curr_base_dir = filepath.Join(config.TargetDirPath, HAS_FACE_FILE_DIR)
				} else {
					log.Printf("NO FACE File: %s", image.File_path)
					curr_count = &no_face_image_cnt
					curr_base_dir = filepath.Join(config.TargetDirPath, NO_FACE_FILE_DIR)
				}
				dir := strconv.Itoa(*curr_count / FILES_PER_DIR)
				destDir := filepath.Join(curr_base_dir, dir)
				if *curr_count%FILES_PER_DIR == 0 {
					created, err := utils.Mkdirp(destDir)
					if err != nil {
						log.Printf("create path error. path: %s, err: %v", destDir, err)
					}
					if created {
						log.Printf("create path OK. path: %s", destDir)
					}
				}
				go createLink(image.File_path, destDir)
				*curr_count++
			} else {
				has_image = false
				break
			}
		}
		wg.Wait()
	}
	iter.Close()
}

func createLink(path string, destDir string) {
	defer wg.Done()
	srcFullPath := filepath.Joins(config.ImageFilePath, path)
	destFilePath := filepath.Join(destDir, filepath.Base(path))
	err := os.Symlink(srcFullPath, destFilePath)
	if err != nil {
		log.Printf("symlink file error. dest path: %s, err: %v", destFilePath, err)
	}
}
