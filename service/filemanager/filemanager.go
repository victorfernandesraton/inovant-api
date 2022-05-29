package responder

import (
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
)

// Uploader service to upload file
type Uploader struct {
	FolderPath string
	AccessURL  string
}

// Run uploader files
func (g *Uploader) Run(file []byte, ext string) (*string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "Error to find working dir")
	}
	workingDir += "/files/"
	if len(workingDir) == 0 {
		workingDir, err := os.Getwd()
		if err != nil {
			return nil, errors.Wrap(err, "Error to find working dir")
		}
		workingDir += "/files/"
		err = createDirIfNotExist(workingDir)
		if err != nil {
			return nil, errors.Wrap(err, "Error to create working dir")
		}
	}
	fileUUID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	fileName := fileUUID.String() + "." + ext
	location := workingDir + fileName
	publicURL := g.AccessURL + "/" + fileName
	err = ioutil.WriteFile(location, file, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "Error saving file")
	}
	return &publicURL, err
}

// createDirIfNotExist create new dir to save nfse files
func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
