package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ubergesundheit/cloudcoverage/imagedimension"
	"github.com/ubergesundheit/cloudcoverage/imageprocessing"
	"github.com/ubergesundheit/cloudcoverage/sensebox"
	"gopkg.in/urfave/cli.v1"
)

func execute(imageDimension imagedimension.ImageDimension, box *sensebox.Sensebox) (err error) {
	//r := raspi.NewAdaptor()
	//r.Connect()
	//tsl45315 := sensors.NewTSL45315Driver(r)

	//err := tsl45315.Start()
	//if err != nil {
	//	fmt.Println(err)
	//}

	//lux, err := tsl45315.LuxTimes(3)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println("Lux", lux)
	location, err := box.QueryLocation()
	if err != nil {
		return
	}

	count, err := imageprocessing.GrabImageAndCountClouds(imageDimension, location, time.Now())
	if err != nil {
		return
	}

	fmt.Println(count)

	return
}

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "half",
		},
		cli.BoolFlag{
			Name: "quarter",
		},
		cli.StringFlag{
			Name: "boxID",
		},
		cli.StringFlag{
			Name: "ccID",
		},
	}

	app.Action = func(c *cli.Context) (err error) {
		var imageDimension = imagedimension.FullSize

		if c.Bool("half") == true && c.Bool("quarter") == true {
			return cli.NewExitError("--half and --quarter are exclusive to each other", 1)
		}

		if c.Bool("half") == true {
			imageDimension = imagedimension.HalfSize
		} else if c.Bool("quarter") == true {
			imageDimension = imagedimension.QuarterSize
		}

		sb, err := sensebox.NewSensebox(c.String("boxID"), c.String("ccID"))
		if err != nil {
			return cli.NewExitError("--boxID and --ccID :"+err.Error(), 1)
		}

		err = execute(imageDimension, sb)
		if err != nil {
			return cli.NewExitError("Image processing execution error: "+err.Error(), 1)
		}

		return nil
	}

	app.Run(os.Args)

}
