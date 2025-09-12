package api

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
)

func TestHandleCreateProperty(t *testing.T) {
	r := chi.NewRouter()
	RegisterRoutes(r)

	srv := httptest.NewServer(r)
	defer srv.Close()

	// Create a request to the /properties endpoint with form data
	data := strings.NewReader("Name=TestProperty&Address=TestAddress&PropertyType=TestType")
	req, err := http.NewRequest("POST", srv.URL+"/properties/new", data)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Could not execute request: %v", err)
	}
	defer res.Body.Close()

	// Check the response status code
	if res.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, res.StatusCode)
	}
}

func TestHandleImportProperty(t *testing.T) {
	r := chi.NewRouter()
	RegisterRoutes(r)

	srv := httptest.NewServer(r)
	defer srv.Close()

	// Create a request to the /properties/import endpoint with a CSV file
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)

	// Create a CSV file
	fileContent := "Name,Address,PropertyType\nTestProperty,TestAddress,TestType"
	file := bytes.NewReader([]byte(fileContent))

	fw, err := w.CreateFormFile("csvFile", "test.csv")
	if err != nil {
		t.Fatalf("Could not create form file: %v", err)
	}
	if _, err := io.Copy(fw, file); err != nil {
		t.Fatalf("Could not copy file content: %v", err)
	}

	w.Close()

	req, err := http.NewRequest("POST", srv.URL+"/properties/import", body)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Execute the request
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Could not execute request: %v", err)
	}
	defer res.Body.Close()

	// Check the response status code
	if res.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected status code %d, got %d", http.StatusSeeOther, res.StatusCode)
	}
}
