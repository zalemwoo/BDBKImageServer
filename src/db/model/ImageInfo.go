package model

import "gopkg.in/mgo.v2/bson"

type ImageInfo struct {
	Id           bson.ObjectId `json:"id" bson:"_id"`
	Src          string        `json:"src"`
	Has_face     bool          `json:"has_face"`
	Face_info    string        `json:"face_info"`
	Face_checked bool          `json:"face_checked"`
	Url          string        `json:"url"`
	Is_cover     bool          `json:"is_cover"`
	Album_url    string        `json:"album_url"`
	Mime         string        `json:"mime"`
	Desc         string        `json:"desc"`
	Width        int           `json:"width"`
	Height       int           `json:"heigh"`
	Size         int           `json:"size"`
	File_name    string        `json:"file_name"`
	File_path    string        `json:"file_path"`
	Person_name  string        `json:"person_name"`
	Person_url   string        `json:"person_url"`
}
