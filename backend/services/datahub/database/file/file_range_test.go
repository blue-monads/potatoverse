package file

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/gin-gonic/gin"
	"github.com/upper/db/v4/adapter/sqlite"
)

func TestStreamFileToHTTP_Range(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup SQLite in-memory
	sess, err := sqlite.Open(sqlite.ConnectionURL{Database: ":memory:"})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer sess.Close()

	// Create tables
	driver := sess.Driver().(*sql.DB)
	_, err = driver.Exec(FileSchemaSQL)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}
	driver.Exec("DELETE FROM FileMeta")
	driver.Exec("DELETE FROM FileBlob")

	fops := NewFileOperations(Options{
		DbSess:           sess,
		MinMultiPartSize: 1024,
		StoreType:        StoreTypeInline,
	})

	ownerID := int64(1)
	content := []byte("Hello World! This is a test file for range support.")
	fileName := "test_inline.txt"

	// Create file
	_, err = fops.CreateFile(ownerID, &datahub.CreateFileRequest{
		Name:      fileName,
		Path:      "",
		CreatedBy: ownerID,
	}, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Test Range Request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/test.txt", nil)
	req.Header.Set("Range", "bytes=0-4")
	c.Request = req

	err = fops.StreamFileToHTTP(ownerID, "", fileName, c)
	if err != nil {
		t.Fatalf("StreamFileToHTTP failed: %v", err)
	}

	if w.Code != http.StatusPartialContent {
		t.Errorf("Expected status 206, got %d", w.Code)
	}

	expected := "Hello"
	if w.Body.String() != expected {
		t.Errorf("Expected body %q, got %q", expected, w.Body.String())
	}

	// Total length is 51
	if w.Header().Get("Content-Range") != "bytes 0-4/51" {
		t.Errorf("Expected Content-Range bytes 0-4/51, got %s", w.Header().Get("Content-Range"))
	}
}

func TestStreamFileToHTTP_Range_External(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir := t.TempDir()

	// Setup SQLite in-memory
	sess, err := sqlite.Open(sqlite.ConnectionURL{Database: ":memory:"})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer sess.Close()

	// Create tables
	driver := sess.Driver().(*sql.DB)
	_, err = driver.Exec(FileSchemaSQL)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}
	driver.Exec("DELETE FROM FileMeta")
	driver.Exec("DELETE FROM FileBlob")

	fops := NewFileOperations(Options{
		DbSess:            sess,
		MinMultiPartSize:  1024,
		StoreType:         StoreTypeExternal,
		ExternalFilesPath: tmpDir,
	})

	ownerID := int64(1)
	content := []byte("External Hello World! Range support test.")
	fileName := "external_test.txt"

	// Create file
	_, err = fops.CreateFile(ownerID, &datahub.CreateFileRequest{
		Name:      fileName,
		Path:      "",
		CreatedBy: ownerID,
	}, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Test Range Request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/external_test.txt", nil)
	req.Header.Set("Range", "bytes=9-13") // "Hello"
	c.Request = req

	err = fops.StreamFileToHTTP(ownerID, "", fileName, c)
	if err != nil {
		t.Fatalf("StreamFileToHTTP failed: %v", err)
	}

	if w.Code != http.StatusPartialContent {
		t.Errorf("Expected status 206, got %d", w.Code)
	}

	expected := "Hello"
	if w.Body.String() != expected {
		t.Errorf("Expected body %q, got %q", expected, w.Body.String())
	}
}

func TestStreamFileToHTTP_Range_Multipart(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup SQLite in-memory
	sess, err := sqlite.Open(sqlite.ConnectionURL{Database: ":memory:"})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer sess.Close()

	// Create tables
	driver := sess.Driver().(*sql.DB)
	_, err = driver.Exec(FileSchemaSQL)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}
	driver.Exec("DELETE FROM FileMeta")
	driver.Exec("DELETE FROM FileBlob")

	fops := NewFileOperations(Options{
		DbSess:           sess,
		MinMultiPartSize: 5, // Small parts for testing
		StoreType:        StoreTypeMultipart,
	})

	ownerID := int64(1)
	content := []byte("Multipart Hello World! Range support test.")
	fileName := "multipart_test.txt"

	// Create file
	_, err = fops.CreateFile(ownerID, &datahub.CreateFileRequest{
		Name:      fileName,
		Path:      "",
		CreatedBy: ownerID,
	}, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Test Range Request
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/multipart_test.txt", nil)
	req.Header.Set("Range", "bytes=10-14") // "Hello"
	c.Request = req

	err = fops.StreamFileToHTTP(ownerID, "", fileName, c)
	if err != nil {
		t.Fatalf("StreamFileToHTTP failed: %v", err)
	}

	if w.Code != http.StatusPartialContent {
		t.Errorf("Expected status 206, got %d", w.Code)
	}

	expected := "Hello"
	if w.Body.String() != expected {
		t.Errorf("Expected body %q, got %q", expected, w.Body.String())
	}
}

func TestStreamFileToHTTP_Range_Multipart_Large(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup SQLite in-memory
	sess, err := sqlite.Open(sqlite.ConnectionURL{Database: ":memory:"})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer sess.Close()

	// Create tables
	driver := sess.Driver().(*sql.DB)
	_, err = driver.Exec(FileSchemaSQL)
	if err != nil {
		t.Fatalf("Failed to create tables: %v", err)
	}
	driver.Exec("DELETE FROM FileMeta")
	driver.Exec("DELETE FROM FileBlob")

	partSize := 1024
	fops := NewFileOperations(Options{
		DbSess:           sess,
		MinMultiPartSize: int64(partSize),
		StoreType:        StoreTypeMultipart,
	})

	ownerID := int64(1)
	// Create a 5KB file (5 parts)
	content := make([]byte, 5120)
	for i := range content {
		content[i] = byte(i % 256)
	}
	fileName := "multipart_large.bin"

	_, err = fops.CreateFile(ownerID, &datahub.CreateFileRequest{
		Name:      fileName,
		Path:      "",
		CreatedBy: ownerID,
	}, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Test Range Request spanning multiple parts (e.g., from part 0 to part 2)
	// Part 0: [0, 1024), Part 1: [1024, 2048), Part 2: [2048, 3072)
	start := 1000
	end := 2100
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest("GET", "/multipart_large.bin", nil)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	c.Request = req

	err = fops.StreamFileToHTTP(ownerID, "", fileName, c)
	if err != nil {
		t.Fatalf("StreamFileToHTTP failed: %v", err)
	}

	if w.Code != http.StatusPartialContent {
		t.Errorf("Expected status 206, got %d", w.Code)
	}

	result := w.Body.Bytes()
	expected := content[start : end+1]
	if !bytes.Equal(result, expected) {
		t.Errorf("Content mismatch, expected len %d, got len %d", len(expected), len(result))
	}
}
