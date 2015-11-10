package model

import "gopkg.in/mgo.v2/bson"

type PersonInfo struct {
	Id          bson.ObjectId `json:"id" bson:"_id"`
	Name        string        `json:"name"`
	Url         string        `json:"url"`
	Tags        []string      `json:"tags"`
	Summary_pic string        `json:"Summary_pic"`
}
