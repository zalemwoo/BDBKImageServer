package facecheck

import (
	"fmt"
	"strconv"
	"strings"
)

type Rect struct {
	Left   int `json:"left"`
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
}

type FaceInfo struct {
	FaceCount int    `json:"facecount"`
	Faces     []Rect `json:"faces"`
	HeadCount int    `json:"headcount"`
	Heads     []Rect `json:"heads"`
}

func str2rect(str string) (rect Rect, err error) {
	defer func() {
		if ex := recover(); ex != nil {
			rect = Rect{}
			err = fmt.Errorf("str2rect error: %v", ex)
		}
	}()

	points := strings.Split(str, ",")
	if len(points) != 4 {
		panic("rect argument format error.")
	}
	left, err := strconv.Atoi(points[0])
	if err != nil {
		panic(err)
	}
	top, err := strconv.Atoi(points[1])
	if err != nil {
		panic(err)
	}
	right, err := strconv.Atoi(points[2])
	if err != nil {
		panic(err)
	}
	bottom, err := strconv.Atoi(points[3])
	if err != nil {
		panic(err)
	}
	rect = Rect{left, top, right, bottom}
	return rect, nil
}

func ParseFaceInfo(message string) (faceinfo *FaceInfo, err error) {
	defer func() {
		if ex := recover(); ex != nil {
			faceinfo = nil
			err = fmt.Errorf("parse face info panic: %v", ex)
		}
	}()

	if len(message) == 0 {
		return nil, nil
	}

	messageParts := strings.Split(message, "|")
	if len(messageParts) != 2 {
		return nil, fmt.Errorf("message format error. message: %s", message)
	}

	faceParts := strings.Split(messageParts[0], ":")[1]
	headParts := strings.Split(messageParts[1], ":")[1]

	if faceParts[len(faceParts)-1] == ';' {
		faceParts = faceParts[:len(faceParts)-1]
	}
	if headParts[len(headParts)-1] == ';' {
		headParts = headParts[:len(headParts)-1]
	}

	facesStr := strings.Split(faceParts, ";")
	headsStr := strings.Split(headParts, ";")

	faceinfo = new(FaceInfo)
	faceinfo.FaceCount = len(facesStr)
	for _, v := range facesStr {
		rect, err := str2rect(v)
		if err == nil {
			faceinfo.Faces = append(faceinfo.Faces, rect)
		}
	}

	faceinfo.HeadCount = len(headsStr)
	for _, v := range headsStr {
		rect, err := str2rect(v)
		if err == nil {
			faceinfo.Heads = append(faceinfo.Heads, rect)
		}
	}
	return
}
