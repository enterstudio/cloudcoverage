package main

import (
	"io"
	"os"

	"github.com/ubergesundheit/cloudcoverage/raspistill"

	"github.com/ubergesundheit/cloudcoverage/sensors"

	"gobot.io/x/gobot/platforms/raspi"

	"github.com/ubergesundheit/cloudcoverage/imagedimension"
	"gopkg.in/urfave/cli.v1"
)

func readLux() (lux float64, err error) {
	r := raspi.NewAdaptor()
	r.Connect()
	tsl45315 := sensors.NewTSL45315Driver(r)

	err = tsl45315.Start()
	if err != nil {
		return
	}

	lux, err = tsl45315.LuxTimes(3)
	if err != nil {
		return
	}
	return
}

func execute(imageDimension imagedimension.ImageDimension, filename string, automaticShutterspeed bool) (err error) {
	if automaticShutterspeed == true {

	}

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
			Name: "output",
		},
		cli.BoolTFlag{
			Name: "automatic-shutterspeed",
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

		output := c.String("output")

		if len(output) == 0 {
			output = "-"
		}

		_, pw := io.Pipe()

		if c.BoolT("automatic-shutterspeed") == true {
			err = raspistill.GrabImage(imageDimension, output, pw)
		} else {
			lux, err := readLux()
			if err != nil {
				return err
			}
			err = raspistill.GrabImageLux(imageDimension, output, lux, pw)
		}
		if err != nil {
			return cli.NewExitError("Image grabbing error: "+err.Error(), 1)
		}

		return nil
	}

	app.Run(os.Args)

}
