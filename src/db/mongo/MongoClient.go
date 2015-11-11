package mongo

import (
	"common"
	"fmt"

	mgo "gopkg.in/mgo.v2"
)

type MongoConfig struct {
	Host     string
	Port     string
	Database string
}

type Mongo struct {
	ServerUrl    string
	DatabaseName string
	Connected    bool
	Session      *mgo.Session
	DB           *mgo.Database
}

func NewMongo(config *MongoConfig) (*Mongo, error) {
	mongo := new(Mongo)
	if err := mongo.init(config); err != nil {
		return nil, err
	}
	return mongo, nil
}

func (self *Mongo) init(config *MongoConfig) error {
	server_addr := config.Host + ":" + config.Port
	dialInfo := new(mgo.DialInfo)
	dialInfo.Addrs = []string{server_addr}
	dialInfo.Database = config.Database
	dialInfo.Timeout = common.MongoConnTimeout

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		return err
	}

	self.ServerUrl = server_addr
	self.DatabaseName = config.Database
	self.Connected = true
	self.Session = session
	self.DB = self.Session.DB(self.DatabaseName)

	return nil
}

func (self *Mongo) Close() {
	self.Session.Close()
}

func (self *Mongo) getCollection(name string) *mgo.Collection {
	if self.Connected == false {
		return nil
	}
	return self.DB.C(name)
}

func (self *Mongo) collectionCount(name string) int {
	coll := self.getCollection(name)
	n, err := coll.Count()
	if err != nil {
		fmt.Println(err)
		return -1
	}
	return n
}
