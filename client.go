package bunnystorage

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/minio/sha256-simd"

	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

type Client struct {
	*resty.Client
	logger resty.Logger
}

// Initialize a new bunnystorage-go client with default settings. Endpoint format is https://<region endpoint>/<Storage Zone Name> e.g. https://la.storage.bunnycdn.com/mystoragezone/
func NewClient(endpoint url.URL, password string) Client {
	return Client{
		resty.
			New().
			SetRetryCount(5).
			AddRetryCondition(
				func(r *resty.Response, err error) bool {
					if err != nil {
						return false
					}
					return r.StatusCode() == http.StatusTooManyRequests
				},
			).
			SetBaseURL(endpoint.String()).
			SetHeader("AccessKey", password),
		logrus.New(),
	}
}

// Add a custom logger. The logger has to implement the resty.Logger interface
func (c *Client) WithLogger(l resty.Logger) *Client {
	c.logger = l
	return c
}

// Uploads a file to the relative path. generateChecksum controls if a checksum gets generated and attached to the upload request. Returns an error.
func (c *Client) Upload(path string, content []byte, generateChecksum bool) error {
	req := c.R().
		SetHeader("Content-Type", "application/octet-stream").
		SetBody(content)

	if generateChecksum {
		checksum := sha256.New()
		_, err := checksum.Write(content)
		if err != nil {
			return err
		}
		hex_checksum := hex.EncodeToString(checksum.Sum(nil))
		req = req.SetHeader("Checksum", hex_checksum)
	}

	resp, err := req.Put(path)

	if err != nil {
		c.logger.Errorf("Put Request Failed: %v", err)
		return err
	}
	if resp.IsError() {
		return errors.New(resp.Status())
	}
	c.logger.Debugf("Put Request Response: %v", resp)

	return nil
}

// Downloads a file from a path.
func (c *Client) Download(path string) ([]byte, error) {
	resp, err := c.R().Get(path)

	if err != nil {
		c.logger.Errorf("Get Request Failed: %v", err)
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(resp.Status())
	}
	c.logger.Debugf("Get Request Response: %v", resp)

	return resp.Body(), nil
}

// Downloads a byte range of a file. Uses the semantics for HTTP range requests
//
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Range_requests
func (c *Client) DownloadPartial(path string, rangeStart int, rangeEnd int) ([]byte, error) {
	rangeHeader := fmt.Sprintf("bytes=%d-%d", rangeStart, rangeEnd)
	resp, err := c.R().
		SetHeader("Range", rangeHeader).
		Get(path)

	if err != nil {
		c.logger.Errorf("Get Range Request Failed: %v", err)
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(resp.Status())
	}
	c.logger.Debugf("Get Range Request Response: %v %v-%v", resp, rangeStart, rangeEnd)

	return resp.Body(), nil
}

// Delete a file or a directory. If the path to delete is a directory, set the isPath flag to true
func (c *Client) Delete(path string, isPath bool) error {
	if isPath {
		path += "/" // The trailing slash is required to delete a directory
	}

	resp, err := c.R().Delete(path)

	if err != nil {
		c.logger.Errorf("Delete Request Failed: %v", err)
		return err
	}
	if resp.IsError() {
		if resp.StatusCode() == http.StatusNotFound {
			return nil // Some clients seem to expect seleting a non-existing file to return without an error
		}
		return errors.New(resp.Status())
	}
	c.logger.Debugf("Delete Request Response: %v", resp)

	return nil
}

// Lists files from a directory.
func (c *Client) List(path string) ([]Object, error) {
	objectList := []Object{}
	resp, err := c.R().
		SetHeader("Accept", "application/json").
		SetResult(&objectList).
		Get(path + "/") // The trailing slash is neccessary, since without it the API will treat the requested directory as a file and returns an empty list

	if err != nil {
		c.logger.Errorf("List Request Failed: %v", err)
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(resp.Status())
	}
	c.logger.Debugf("List Request Response: %v", resp)

	return objectList, nil
}

// Describes an Object. EXPERIMENTAL. The official Java SDK uses it, but the DESCRIBE HTTP method used is not officially documented.
func (c *Client) Describe(path string) (Object, error) {
	object := Object{}

	resp, err := c.R().
		SetHeader("Accept", "application/json").
		SetResult(&object).
		Execute("DESCRIBE", path)

	if err != nil {
		c.logger.Errorf("Describe Request Failed: %v", err)
		return object, err
	}
	if resp.IsError() {
		return object, errors.New(resp.Status())
	}
	c.logger.Debugf("Describe Request Response: %v", resp)

	return object, nil
}
