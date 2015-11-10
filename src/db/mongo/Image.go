package mongo

import (
	"db/model"
	"log"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const IMAGE_INFO_COLLECTION_NAME = "image_info"

func (self *Mongo) imageCollection() *mgo.Collection {
	return self.getCollection(IMAGE_INFO_COLLECTION_NAME)
}

func (self *Mongo) ImageCount() int {
	return self.collectionCount(IMAGE_INFO_COLLECTION_NAME)
}

func (self *Mongo) ImageAll() *mgo.Iter {
	return self.getCollection(IMAGE_INFO_COLLECTION_NAME).Find(nil).Iter()
}

func (self *Mongo) ImageFaceChecked(checked bool) *mgo.Iter {
	return self.getCollection(IMAGE_INFO_COLLECTION_NAME).Find(bson.M{"face_checked": checked}).Iter()
}

func (self *Mongo) ImageByName(name string) *mgo.Iter {
	return self.getCollection(IMAGE_INFO_COLLECTION_NAME).Find(bson.M{"file_name": name}).Iter()
}

func (self *Mongo) ImageUpdateHasFace(image *model.ImageInfo) {
	err := self.getCollection(IMAGE_INFO_COLLECTION_NAME).UpdateId(image.Id, bson.M{"$set": bson.M{"has_face": image.Has_face, "face_info": image.Face_info, "face_checked": image.Face_checked}})
	if err != nil {
		log.Printf("Mongo::Image Update Error. err: %v", err)
	}
}
