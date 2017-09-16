package imageprocessing

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	suncalc "github.com/mourner/suncalc-go"
	"github.com/ubergesundheit/cloudcoverage/sensebox"
)

var (
	staticRaspistillArgs       = []string{"-o", "-", "--timeout", "1", "--thumb", "none", "--nopreview", "--quality", "100", "--metering", "backlit", "--ISO", "100", "--awb", "sun"}
	staticConvertArgs          = []string{"jpg:-", "-distort", "barrel", "0.005 -0.025 -0.028", "-fx", "(b==0) ? 0 : (r/b)"}
	staticConvertHistogramArgs = []string{"-define", "histogram:unique-colors=true", "-format", "%c", "histogram:info:-"}
	FullSize                   = ImageDimension{2592, 1944}
	HalfSize                   = ImageDimension{FullSize.width / 2, FullSize.height / 2}
	QuarterSize                = ImageDimension{HalfSize.width / 2, HalfSize.height / 2}
)

func bla(winkel, laenge, centerX, centerY float64) (x, y float64) {
	x = (math.Sin((winkel * 180 / math.Pi)) * laenge) + centerX
	y = (math.Cos((winkel * 180 / math.Pi)) * laenge) + centerY
	return
}

func convertFillArgs(imageDimension ImageDimension, location *sensebox.Location, timestamp time.Time) []string {
	azimuth, altitude := suncalc.SunPosition(timestamp, location.Latitude, location.Longitude)

	// var sunPos = SunCalc.getPosition(new Date(req.body.timestamp), location.lat, location.lon)
	// sunPos.azimuth *= (180 / Math.PI)
	// sunPos.azimuth += 180

	// // factor 0.7 for a big kegel
	// sunPos.altitude = sunPos.altitude * (180 / Math.PI) * 0.7;

	// // center of the image
	// var centerx = Math.floor(image.bitmap.width / 2)
	// var centery = Math.floor(image.bitmap.height / 2)

	// // min and max for the kegel
	// var azimin = ((sunPos.azimuth - sunPos.altitude) < 0 ? 360 - Math.abs(sunPos.azimuth - sunPos.altitude) : sunPos.azimuth - sunPos.altitude);
	// var azimax = (sunPos.azimuth + sunPos.altitude) > 360 ? 360 - sunPos.azimuth + sunPos.altitude : sunPos.azimuth + sunPos.altitude;
	altitude *= (180 / math.Pi) * 0.7

	if altitude < 15 {
		return []string{}
	}

	azimuth *= (180 / math.Pi)
	azimuth += 180

	azimin := azimuth - altitude
	if azimin < 0 {
		azimin = 360 - math.Abs(azimuth-altitude)
	}
	azimax := azimuth + altitude
	if azimax > 360 {
		azimax = 360 - azimuth + altitude
	}

	fmt.Println(azimuth, altitude, azimin, azimax)
	centerX, centerY := imageDimension.center()

	// b := (yA - m * xA)
	//b1 := centerY - math.Tan(azimin)*centerX
	//b2 := centerY - math.Tan(azimax)*centerX

	x1, y1 := bla(azimax, 5000, centerX, centerY)
	x2, y2 := bla(azimin, 5000, centerX, centerY)

	// polygon := fmt.Sprintf("polygon %.2f,%.2f %.2f,%.2f %.2f,%.2f", 0.0, 0.0, centerX, centerY, 100.0, 100.0)
	path := fmt.Sprintf("path 'M %.2f,%.2f L %.2f,%.2f L %.2f,%.2f Z'", centerX, centerY, x1, y1, x2, y2)

	return []string{"-fill", `#f0f`, "-draw", path}
}

func GrabImageAndCountClouds(imageDimension ImageDimension, location *sensebox.Location, timestamp time.Time) (count int, err error) {
	raspistillArgs := append([]string{"--width", strconv.Itoa(imageDimension.width), "--height", strconv.Itoa(imageDimension.height)}, staticRaspistillArgs...)
	fmt.Println(raspistillArgs)
	raspistillCommand := exec.Command("raspistill", raspistillArgs...)

	convertArgs := append(staticConvertArgs, convertFillArgs(imageDimension, location, timestamp)...)
	// convertArgs = append(convertArgs, staticConvertHistogramArgs...)
	convertArgs = append(convertArgs, "img4.jpg")
	fmt.Println(convertArgs)
	convertCommand := exec.Command("convert", convertArgs...)

	pr, pw := io.Pipe()
	raspistillCommand.Stdout = pw
	convertCommand.Stdin = pr

	var pixelCount bytes.Buffer
	convertCommand.Stdout = &pixelCount

	var raspistillStdErr bytes.Buffer
	var convertStdErr bytes.Buffer
	raspistillCommand.Stderr = &raspistillStdErr
	convertCommand.Stderr = &convertStdErr

	err = raspistillCommand.Start()
	if err != nil {
		return
	}

	err = convertCommand.Start()
	if err != nil {
		return
	}

	go func() {
		defer pw.Close()

		raspistillCommand.Wait()
	}()
	convertCommand.Wait()
	if len(raspistillStdErr.String()) != 0 {
		err = fmt.Errorf("raspistill error: %q with arguments %q", strings.TrimSpace(raspistillStdErr.String()), raspistillArgs)
		return
	}

	if len(convertStdErr.String()) != 0 {
		err = fmt.Errorf("convert error: %q with arguments %q", strings.TrimSpace(convertStdErr.String()), convertArgs)
		return
	}

	count, err = strconv.Atoi(strings.TrimSpace(pixelCount.String()))
	if err != nil {
		return
	}

	return
}
