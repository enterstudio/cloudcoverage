package raspistill

import (
	"bytes"
	"fmt"
	"io"
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

func GrabImage(imageDimension imagedimension.ImageDimension, filename string, stdout *io.PipeWriter) (err error) {
	raspistillArgs := append([]string{
		"--width", strconv.Itoa(imageDimension.Width()),
		"--height", strconv.Itoa(imageDimension.Height()),
		"--output", filename,
	}, staticRaspistillArgs...)

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
