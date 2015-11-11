package main

import (
	"encoding/json"
	"facecheck"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"strconv"
	"utils"

	"github.com/codegangsta/cli"
)

var app *cli.App

func main() {
	app = cli.NewApp()
	app.Name = "facecrop"
	app.Usage = "crop face from image by rect info arguments"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "path, p",
			Value: "",
			Usage: "path of image file.",
		},
		cli.StringFlag{
			Name:  "rect, r",
			Value: "",
			Usage: "rect to crop. (format: left,top,right,bottom)",
		},
		cli.StringFlag{
			Name:  "out, o",
			Value: "out",
			Usage: "path of output file.",
		},
	}

	app.Action = func(c *cli.Context) {
		pathArg := c.String("path")
		rectArg := c.String("rect")
		outArg := c.String("out")

		if len(pathArg) == 0 || len(rectArg) == 0 {
			println("path and rect are required.")
			cli.ShowAppHelp(c)
			return
		}

		fmt.Printf("Processing image file path: %s\n", pathArg)

		file_in, err := os.Open(pathArg)
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

		faceInfo, err := facecheck.ParseFaceInfo(rectArg)
		if err != nil {
			fmt.Printf("Parse Rect error. err: %v, rect: %s\n", err, rectArg)
		}

		faceInfo_json, _ := json.Marshal(faceInfo)
		fmt.Printf("Face info is: %s\n", faceInfo_json)

		writeImageFunc := func(r *facecheck.Rect, suffix string, i int) {
			fmt.Printf("%s rect is: %v\n", suffix, r)
			rect := image.Rect(r.Left, r.Top, r.Right, r.Bottom)
			file_out, err := os.Create(outArg + "_" + suffix + strconv.Itoa(i) + ".jpg")
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

		fmt.Println("DONE!")
	}

	app.Run(os.Args)
}
