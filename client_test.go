package bunnystorage_test

import (
	"crypto/rand"
	"net/url"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/google/uuid"
	bunnystorage "github.com/l0wl3vel/bunny-storage-go-sdk"
)

var password string
var endpoint url.URL

var bunnyclient bunnystorage.Client

var testingDirectory string

func TestMain(m *testing.M) {
	password = os.Getenv("BUNNY_PASSWORD")
	endpointString := os.Getenv("BUNNY_ENDPOINT")
	if password == "" {
		panic("No API Key provided (BUNNY_PASSWORD)")
	}
	if endpointString == "" {
		panic("No Endpoint provided (BUNNY_ENDPOINT)")
	}
	endpoint, err := endpoint.Parse(endpointString)
	if err != nil {
		panic(err)
	}

	testingDirectory = uuid.New().String()

	bunnyclient = bunnystorage.NewClient(*endpoint, password)
	m.Run()
}

func TestRangeDownload(t *testing.T) {
	t.Cleanup(func() { DeleteTestPath(t) })

	content, filepath := UploadRandomFile1MB(t)

	contentRange, err := bunnyclient.DownloadPartial(filepath, 20, 1023)
	if err != nil {
		t.Error(err)
	}
	// Go uses [inclusive:exclusive] slice indexing and HTTP Ranges use [inclusive:inclusive]
	expected_range := content[20:1024]
	if !reflect.DeepEqual(contentRange, expected_range) {
		t.Errorf("Ranged download did not return the expected range len_input:%v len_test:%v", len(expected_range), len(contentRange))
	}
}

func TestDescribeFile(t *testing.T) {
	t.Cleanup(func() { DeleteTestPath(t) })
	_, name := UploadRandomFile1MB(t)
	obj, err := bunnyclient.Describe(name)
	if err != nil {
		t.Error(err)
	}
	empty_object := bunnystorage.Object{}
	if obj == empty_object {
		t.Errorf("Returned empty DESCRIBE API response")
	}
}

func DeleteFileWithPathSetToTrue(t *testing.T) {
	t.Cleanup(func() { DeleteTestPath(t) })
	_, testpath := UploadRandomFile1MB(t)

	err := bunnyclient.Delete(testpath, true)
	if err != nil {
		t.Error(err)
	}
	list := ListFilesInTestDir(t)
	if len(list) != 0 {
		t.Errorf("Returned List not as long as expected: Got: %v Expected %v", len(list), 1)
	}
}

func TestGetNonexistentFile(t *testing.T) {
	body, err := bunnyclient.Download("thispathdoesnotexist")
	if err == nil {
		t.Error("Error should not be nil when getting a non-existent file")
	}
	if body != nil {
		t.Error("Returned buffer should be nil when getting a non-existent file")
	}
}

func TestDeleteNonexistentFile(t *testing.T) {
	err := bunnyclient.Delete("thispathdoesnotexist", false)
	if err != nil {
		t.Error("Expected no error when deleting a file that does not exist")
	}
}

func TestDownloadAfterUpload1M(t *testing.T) {
	t.Cleanup(func() { DeleteTestPath(t) })
	input, _ := UploadRandomFile1MB(t) // 1MB file size
	output := DownloadFile(t)
	list := ListFilesInTestDir(t)
	if len(list) != 1 {
		t.Errorf("Returned List not as long as expected: Got: %v Expected %v", len(list), 1)
	}

	if !reflect.DeepEqual(input, output) {
		t.Error("Downloaded Content does not match uploaded content")
	}
}

func TestListAfterUploadWithExtraTrailingSlash(t *testing.T) {
	t.Cleanup(func() { DeleteTestPath(t) })
	_, _ = UploadRandomFile1MB(t) // 1MB file size
	list, err := bunnyclient.List(testingDirectory + "/")
	if err != nil {
		t.Error(err)
	}
	if len(list) != 1 {
		t.Errorf("Returned List not as long as expected: Got: %v Expected %v", len(list), 1)
	}
}

func TestListOnMissingDirectory(t *testing.T) {
	list, err := bunnyclient.List(testingDirectory + t.Name())
	if err != nil {
		t.Error(err)
	}
	if len(list) != 0 {
		t.Errorf("Returned List not as long as expected: Got: %v Expected %v", len(list), 1)
	}
}

func ListFilesInTestDir(t *testing.T) []bunnystorage.Object {
	items, err := bunnyclient.List(testingDirectory)
	if err != nil {
		t.Error(err)
	}
	return items
}

func DownloadFile(t *testing.T) []byte {
	t.Helper()
	testpath := path.Join(testingDirectory, t.Name())

	body, err := bunnyclient.Download(testpath)
	if err != nil {
		t.Error(err)
	}
	return body
}

func UploadRandomFile1MB(t *testing.T) ([]byte, string) {
	t.Helper()
	testpath := path.Join(testingDirectory, t.Name())
	testcontent := make([]byte, 1048576)

	_, err := rand.Read(testcontent)
	if err != nil {
		t.Error(err)
	}

	err = bunnyclient.Upload(testpath, testcontent, true)
	if err != nil {
		t.Error(err)
	}
	return testcontent, testpath
}

func DeleteFile(t *testing.T) {
	t.Helper()
	testpath := path.Join(testingDirectory, t.Name())

	err := bunnyclient.Delete(testpath, false)
	if err != nil {
		t.Error(err)
	}
}

func DeleteTestPath(t *testing.T) {
	t.Helper()
	testpath := path.Join(testingDirectory)

	err := bunnyclient.Delete(testpath, true)
	if err != nil {
		t.Error(err)
	}
}
