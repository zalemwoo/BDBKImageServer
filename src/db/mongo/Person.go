package mongo

import mgo "gopkg.in/mgo.v2"

const PERSON_INFO_COLLECTION_NAME = "person_info"

func (self *Mongo) personCollection() *mgo.Collection {
	return self.getCollection(PERSON_INFO_COLLECTION_NAME)
}

func (self *Mongo) PersonCount() int {
	return self.collectionCount(PERSON_INFO_COLLECTION_NAME)
}

func (self *Mongo) PersonAll() *mgo.Iter {
	return self.getCollection(PERSON_INFO_COLLECTION_NAME).Find(nil).Iter()
}
