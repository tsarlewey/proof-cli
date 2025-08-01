package real_estate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/tsarlewey/proof-cli/pkg/utils"
)

const (
	// DocumentsEndpoint is the base endpoint for documents API
	DocumentsEndpoint = "/mortgage/v2/documents"
)

// Document represents a document in the real estate system
type Document struct {
	ID               string                 `json:"id,omitempty"`
	FileName         string                 `json:"file_name,omitempty"`
	ContentType      string                 `json:"content_type,omitempty"`
	DateCreated      *time.Time             `json:"date_created,omitempty"`
	DateUpdated      *time.Time             `json:"date_updated,omitempty"`
	URL              string                 `json:"url,omitempty"`
	ThumbnailURL     string                 `json:"thumbnail_url,omitempty"`
	Status           string                 `json:"status,omitempty"`
	UploadStatus     string                 `json:"upload_status,omitempty"`
	ProcessingStatus string                 `json:"processing_status,omitempty"`
	PageCount        int                    `json:"page_count,omitempty"`
	FileSize         int64                  `json:"file_size,omitempty"`
	TransactionID    string                 `json:"transaction_id,omitempty"`
	ExternalID       string                 `json:"external_id,omitempty"`
	DocumentType     string                 `json:"document_type,omitempty"`
	Tags             []string               `json:"tags,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// AddDocumentRequest represents a request to add a document to a transaction
type AddDocumentRequest struct {
	File         string   `json:"file,omitempty"`          // File path for upload
	FileName     string   `json:"file_name,omitempty"`     // Name of the file
	ExternalID   string   `json:"external_id,omitempty"`   // External identifier
	DocumentType string   `json:"document_type,omitempty"` // Type of document
	Tags         []string `json:"tags,omitempty"`          // Tags for categorization
}

// AddExternalDocumentRequest represents a request to add an external document
type AddExternalDocumentRequest struct {
	URL          string   `json:"url"`                     // URL of the external document
	FileName     string   `json:"file_name,omitempty"`     // Name of the file
	ExternalID   string   `json:"external_id,omitempty"`   // External identifier
	DocumentType string   `json:"document_type,omitempty"` // Type of document
	Tags         []string `json:"tags,omitempty"`          // Tags for categorization
}

// ListDocumentsParams represents query parameters for listing documents
type ListDocumentsParams struct {
	Limit         int    `json:"limit,omitempty"`
	Offset        int    `json:"offset,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	DocumentType  string `json:"document_type,omitempty"`
	Status        string `json:"status,omitempty"`
}

// ListDocumentsResponse represents the response from listing documents
type ListDocumentsResponse struct {
	Count    int        `json:"count"`
	Next     string     `json:"next,omitempty"`
	Previous string     `json:"previous,omitempty"`
	Results  []Document `json:"results"`
}

// AddDocument adds a document to a transaction by uploading a file
func AddDocument(client *utils.ProofClient, transactionID string, filePath string, request *AddDocumentRequest) ([]byte, error) {
	// For file uploads, we need to construct the full URL and handle multipart ourselves
	config, err := utils.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	endpoint := config.APIEndpoint + TransactionsEndpoint + "/" + transactionID + "/documents"

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Create a buffer to write our multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the file
	fileName := filepath.Base(filePath)
	if request != nil && request.FileName != "" {
		fileName = request.FileName
	}

	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// Add other fields if provided
	if request != nil {
		if request.ExternalID != "" {
			writer.WriteField("external_id", request.ExternalID)
		}
		if request.DocumentType != "" {
			writer.WriteField("document_type", request.DocumentType)
		}
		if len(request.Tags) > 0 {
			tagsJSON, _ := json.Marshal(request.Tags)
			writer.WriteField("tags", string(tagsJSON))
		}
	}

	// Close the writer
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// Create the request
	req, err := http.NewRequest("POST", endpoint, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Get API key from client
	apiKey, err := utils.GetAPIKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	// Set headers
	req.Header.Set("ApiKey", apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: time.Duration(fileInfo.Size()/1024/10+30) * time.Second, // Dynamic timeout based on file size
	}

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return respBody, nil
}

// AddExternalDocument adds an external document to a transaction by URL
func AddExternalDocument(client *utils.ProofClient, transactionID string, request *AddExternalDocumentRequest) ([]byte, error) {
	path := TransactionsEndpoint + "/" + transactionID + "/documents/external"
	return client.Post(path, request)
}

// GetDocument retrieves a specific document by ID
func GetDocument(client *utils.ProofClient, documentID string) ([]byte, error) {
	path := DocumentsEndpoint + "/" + documentID
	return client.Get(path)
}

// DeleteDocument deletes a document from a transaction
func DeleteDocument(client *utils.ProofClient, transactionID string, documentID string) ([]byte, error) {
	path := TransactionsEndpoint + "/" + transactionID + "/documents/" + documentID
	return client.Delete(path)
}

// ListDocuments retrieves a list of documents
func ListDocuments(client *utils.ProofClient, params *ListDocumentsParams) ([]byte, error) {
	path := DocumentsEndpoint

	// Build query parameters
	queryParams := utils.BuildQueryParams(params)
	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	return client.Get(path)
}

// DownloadDocument downloads a document to a local file
func DownloadDocument(apiKey string, documentURL string, outputPath string) error {
	// Create the request
	req, err := http.NewRequest("GET", documentURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("ApiKey", apiKey)

	// Create HTTP client
	client := &http.Client{
		Timeout: 5 * time.Minute, // 5 minute timeout for downloads
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Create the output file
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer out.Close()

	// Copy the response body to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
