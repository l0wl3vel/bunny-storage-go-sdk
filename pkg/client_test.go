package bunnystorage_test

import (
	"crypto/rand"
	"net/url"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/google/uuid"
	bunnystorage "github.com/l0wl3vel/bunny-storage-go-sdk/pkg"
)

var password string
var endpoint url.URL

var bunnyclient bunnystorage.Client

var testingDirectory string

func TestMain(m *testing.M)	{
	password = os.Getenv("BUNNY_PASSWORD")
	endpointString := os.Getenv("BUNNY_ENDPOINT")
	if password == ""	{
		panic("No API Key provided (BUNNY_PASSWORD)")
	}
	if endpointString == ""	{
		panic("No Endpoint provided (BUNNY_ENDPOINT)")
	}
	endpoint, err := endpoint.Parse(endpointString)
	if err != nil	{
		panic(err)
	}

	testingDirectory = uuid.New().String()

	bunnyclient = bunnystorage.NewClient(*endpoint, password)
	m.Run()
}

func DeleteFileWithPathSetToTrue(t *testing.T)	{
	_ = UploadRandomFile1MB(t)
	testpath := path.Join(testingDirectory, t.Name())

	err := bunnyclient.Delete(testpath, true)
	if err != nil	{
		t.Error(err)
	}
	list := ListFilesInTestDir(t)
	if len(list) != 0	{
		t.Errorf("Returned List not as long as expected: Got: %v Expected %v", len(list), 1)
	}
}

func TestDownloadAfterUpload1M(t *testing.T)	{
	input := UploadRandomFile1MB(t) // 1MB file size
	output := DownloadFile(t)
	list := ListFilesInTestDir(t)
	if len(list) != 1	{
		t.Errorf("Returned List not as long as expected: Got: %v Expected %v", len(list), 1)
	}
	DeleteTestPath(t)

	if !reflect.DeepEqual(input, output)	{
		t.Error("Downloaded Content does not match uploaded content")
	}
}

func TestListAfterUploadWithExtraTrailingSlash(t *testing.T)	{
	_ = UploadRandomFile1MB(t) // 1MB file size
	list, err := bunnyclient.List(testingDirectory+"/")
	if err != nil	{
		t.Error(err)
	}
	if len(list) != 1	{
		t.Errorf("Returned List not as long as expected: Got: %v Expected %v", len(list), 1)
	}
	DeleteTestPath(t)
}

func TestListOnMissingDirectory(t *testing.T)	{
	list, err := bunnyclient.List(testingDirectory+t.Name())
	if err != nil	{
		t.Error(err)
	}
	if len(list) != 0	{
		t.Errorf("Returned List not as long as expected: Got: %v Expected %v", len(list), 1)
	}
}

func ListFilesInTestDir(t *testing.T)	[]bunnystorage.Object	{
	items, err := bunnyclient.List(testingDirectory)
	if err != nil	{
		t.Error(err)
	}
	return items
}

func DownloadFile(t *testing.T) []byte	{
	t.Helper()
	testpath := path.Join(testingDirectory, t.Name())

	body, err := bunnyclient.Download(testpath)
	if err != nil	{
		t.Error(err)
	}
	return body
}

func UploadRandomFile1MB(t *testing.T) []byte	{
	t.Helper()
	testpath := path.Join(testingDirectory, t.Name())
	testcontent := make([]byte, 1048576)

	t.Cleanup(func() {
		bunnyclient.Delete(testpath, false)
	})

	_, err := rand.Read(testcontent)
	if err != nil 	{
		t.Error(err)
	}

	err = bunnyclient.Upload(testpath, testcontent, true)
	if err != nil 	{
		t.Error(err)
	}
	return testcontent
}

func DeleteFile(t *testing.T)	{
	t.Helper()
	testpath := path.Join(testingDirectory, t.Name())

	err := bunnyclient.Delete(testpath, false)
	if err != nil	{
		t.Error(err)
	}
}

func DeleteTestPath(t *testing.T)	{
	t.Helper()
	testpath := path.Join(testingDirectory)

	err := bunnyclient.Delete(testpath, true)
	if err != nil	{
		t.Error(err)
	}
}