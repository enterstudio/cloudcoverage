package sensebox

import (
	"bytes"
	"encoding/hex"
	"errors"
	"html/template"
)

const (
	// openSenseMapAPIBoxURLTemplate         = "https://api.opensensemap.org/boxes/{{.BoxID}}"
	// openSenseMapAPIValueSubmitURLTemplate = "https://api.opensensemap.org/boxes/{{.BoxID}}/{{.CloudcoverageID}}"
	openSenseMapAPIBoxURLTemplate         = "https://api.osem.vo1d.space/boxes/{{.BoxID}}"
	openSenseMapAPIValueSubmitURLTemplate = "https://api.osem.vo1d.space/boxes/{{.BoxID}}/{{.CloudcoverageID}}"
)

type Sensebox struct {
	BoxID, CloudcoverageID, boxURL, valueSubmitURL string
}

func validateID(id string) (err error) {
	hex, err := hex.DecodeString(id)
	if err != nil {
		return
	}

	if len(hex) != 12 {
		return errors.New("id must be exactly 24 characters long")
	}

	return
}

func (b *Sensebox) prepareTemplate(templateString, templateName string) (templated string, err error) {
	tpl, err := template.New(templateName).Parse(openSenseMapAPIBoxURLTemplate)
	if err != nil {
		return
	}
	var tplBuffer bytes.Buffer
	tpl.Execute(&tplBuffer, b)

	templated = tplBuffer.String()
	return
}

func NewSensebox(boxID, cloudcoverageID string) (senseBox *Sensebox, err error) {
	err = validateID(boxID)
	if err != nil {
		return
	}
	err = validateID(cloudcoverageID)
	if err != nil {
		return
	}

	senseBox = &Sensebox{BoxID: boxID, CloudcoverageID: cloudcoverageID}

	senseBox.boxURL, err = senseBox.prepareTemplate(openSenseMapAPIBoxURLTemplate, "boxURL")
	if err != nil {
		return
	}

	senseBox.valueSubmitURL, err = senseBox.prepareTemplate(openSenseMapAPIValueSubmitURLTemplate, "valueSubmitURL")
	if err != nil {
		return
	}

	return
}
