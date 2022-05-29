package handler

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

// FileHandler service to create handler
type FileHandler struct {
	upload   func(file []byte, ext string) (*string, error)
	serveDir string
}

// Get returns an echo handler
// @Summary file.Get
// @Description Get file
// @Produce  application/octet-stream
// @Param context query string false "Context to return"
// @Success 200 {object} handler.fileResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /files/{file} [get]
func (handler *FileHandler) Get(c echo.Context) error {
	file := c.Param("file")
	workingDir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Error to find working dir")
	}
	workingDir += "/files/"
	location := workingDir + file
	// location := handler.serveDir + "/" + file
	return c.File(location)
}

// GetTemplateImages returns an echo handler
// @Summary file.GetTemplateImages
// @Description GetTemplateImages file
// @Produce  application/octet-stream
// @Param context query string false "Context to return"
// @Success 200 {object} string
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /templates/images/{file} [get]
func (handler *FileHandler) GetTemplateImages(c echo.Context) error {
	file := c.Param("file")
	workingDir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, "Error to find working dir")
	}
	workingDir += "/templates/images/"
	location := workingDir + file
	return c.File(location)
}

// Upload digital certificate file returns an echo handler
// @Summary certificate.Upload
// @Description Upload digital certificate file
// @Accept  json
// @Produce  json
// @Param context query string false "Context to return"
// @Param formdata body string true "Update"
// @Success 200 {object} handler.fileResponse
// @Failure 400 {object} handler.errorResponse
// @Failure 404 {object} handler.errorResponse
// @Failure 500 {object} handler.errorResponse
// @Router /files/ [post]
func (handler *FileHandler) Upload(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	ext := strings.Split(file.Filename, ".")[1]
	data := bytes.NewBuffer(nil)
	// Copy
	if _, err = io.Copy(data, src); err != nil {
		return errors.Wrap(err, "Failed copy file")
	}
	bdata, err := ioutil.ReadAll(data)
	if err != nil {
		return errors.Wrap(err, "Failed convert file")
	}
	u, err := handler.upload([]byte(bdata), ext)
	if err != nil {
		return errors.Wrap(err, "Failed to create new file")
	}

	return c.JSON(http.StatusOK, fileResponse{
		dataResponse: dataResponse{
			Context: c.QueryParam("context"),
		},
		Data: fileUpload{
			singleItemData: singleItemData{dataDetail: dataDetail{Kind: "Filename"}},
			Item:           fileResult{Filename: *u},
		},
	})
}

type fileResponse struct {
	dataResponse
	Data fileUpload `json:"data"`
}

type fileUpload struct {
	singleItemData
	Item fileResult `json:"item"`
}

type fileResult struct {
	Filename string `json:"filename"`
}
