package raspistill

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strconv"
	"strings"

	"github.com/ubergesundheit/cloudcoverage/imagedimension"
)

var staticRaspistillArgs = []string{
	"--nopreview",
	"--timeout", "1",
	"--thumb", "none",
	"--quality", "100",
	"--metering", "backlit",
	"--ISO", "100",
	"--awb", "sun",
}

func grabImage(raspistillArgs []string, stdout *io.PipeWriter) (err error) {
	fmt.Println(raspistillArgs)
	raspistillCommand := exec.Command("raspistill", raspistillArgs...)

	var raspistillStdErrBuffer bytes.Buffer
	raspistillCommand.Stdout = stdout
	raspistillCommand.Stderr = &raspistillStdErrBuffer

	err = raspistillCommand.Start()
	if err != nil {
		return
	}

	if len(raspistillStdErrBuffer.String()) != 0 {
		err = fmt.Errorf("raspistill error: %q with arguments %q", strings.TrimSpace(raspistillStdErrBuffer.String()), raspistillArgs)
		return
	}

	go func() {
		raspistillCommand.Wait()
	}()

	return
}

func widthHeightOutput(imageDimension imagedimension.ImageDimension, filename string) []string {
	return append([]string{
		"--width", strconv.Itoa(imageDimension.Width()),
		"--height", strconv.Itoa(imageDimension.Height()),
		"--output", filename,
	}, staticRaspistillArgs...)
}

func calculateShutterspeed(lux float64) float64 {
	if lux == 0 {
		return 6000000000
	}

	return (0.7368 * math.Pow(lux, -0.915)) * 1000000
}

func GrabImage(imageDimension imagedimension.ImageDimension, filename string, stdout *io.PipeWriter) (err error) {
	err = grabImage(widthHeightOutput(imageDimension, filename), stdout)
	if err != nil {
		return
	}

	return
}

func GrabImageLux(imageDimension imagedimension.ImageDimension, filename string, lux float64, stdout *io.PipeWriter) (err error) {
	args := widthHeightOutput(imageDimension, filename)
	args = append(args,
		"--shutter", strconv.FormatFloat(calculateShutterspeed(lux), 'f', -1, 64),
	)

	err = grabImage(args, stdout)
	if err != nil {
		return
	}

	return
}
