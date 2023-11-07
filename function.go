package function

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/vmihailenco/msgpack" // Import MessagePack package

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.HTTP("JoinChunks", joinChunks)
}

type RequestData struct {
	Hash          string `msgpack:"hash"`            // Use MessagePack tag
	HowManyChunks int    `msgpack:"how_many_chunks"` // Use MessagePack tag
}

type ResponseData struct {
	Hash string `msgpack:"hash"` // Use MessagePack tag
}

type ErrorData struct {
	Message string `msgpack:"message"` // Use MessagePack tag for error message
}

func sendError(w http.ResponseWriter, message string) {
	errorData := &ErrorData{Message: message}

	// Serialize the error response to MessagePack
	errorBytes, err := msgpack.Marshal(errorData)
	if err != nil {
		log.Printf("Failed to marshal error data: %v", err)
		http.Error(w, "Failed to create error response", http.StatusInternalServerError)
		return
	}

	// Respond with the MessagePack error response
	w.Header().Set("Content-Type", "application/msgpack")
	w.WriteHeader(http.StatusBadRequest) // Use an appropriate HTTP status code
	w.Write(errorBytes)
}

func joinChunks(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Parse the request data using MessagePack
	requestData := &RequestData{}
	decoder := msgpack.NewDecoder(r.Body)
	if err := decoder.Decode(requestData); err != nil {
		sendError(w, "Failed to parse request data")
		return
	}

	// Initialize the Google Cloud Storage client
	client, err := storage.NewClient(ctx)
	if err != nil {
		sendError(w, "Failed to create client")
		return
	}
	defer client.Close()

	// Get a handle to your bucket
	bucket := client.Bucket("your-bucket-name") // Replace with your bucket name

	// Create a channel to collect chunk data
	chunkDataChan := make(chan []byte)
	defer close(chunkDataChan)

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	for i := 0; i < int(requestData.HowManyChunks); i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Read each chunk from the bucket
			chunkName := fmt.Sprintf("%s.%d", requestData.Hash, i)
			rc, err := bucket.Object(chunkName).NewReader(ctx)
			if err != nil {
				sendError(w, "Failed to read chunk")
				return
			}
			defer rc.Close()

			chunkData, err := io.ReadAll(rc)
			if err != nil {
				sendError(w, "Failed to read chunk")
				return
			}

			// Send the chunk data to the channel
			chunkDataChan <- chunkData
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close the channel to signal that all chunks are received
	close(chunkDataChan)

	// Collect the chunk data from the channel and join it
	joinedContent := []byte{}
	for chunkData := range chunkDataChan {
		joinedContent = append(joinedContent, chunkData...)
	}

	// Calculate the MD5 hash of the joined content
	md5Sum := md5Sum(joinedContent)

	// Calculate the size of the joined content
	size := len(joinedContent)

	// Define the MIME type based on your requirements (e.g., application/octet-stream)
	mimeType := "application/octet-stream"

	// Write the joined content to a new file
	outputObject := bucket.Object(requestData.Hash)
	wc := outputObject.NewWriter(ctx)
	if _, err := io.Copy(wc, strings.NewReader(string(joinedContent))); err != nil {
		sendError(w, "Failed to write joined content")
		return
	}
	wc.Close()

	// Create a .meta file with the MD5 hash, size, and MIME type
	metaData := fmt.Sprintf(`{"hash":"%s","size":%d,"mime":"%s"}`, md5Sum, size, mimeType)
	metaObject := bucket.Object(fmt.Sprintf("%s.meta", requestData.Hash))
	wc = metaObject.NewWriter(ctx)
	if _, err := wc.Write([]byte(metaData)); err != nil {
		sendError(w, "Failed to write .meta content")
		return
	}
	wc.Close()

	// Create a MessagePack response
	responseData := &ResponseData{Hash: md5Sum}

	// Serialize the response to MessagePack
	responseBytes, err := msgpack.Marshal(responseData)
	if err != nil {
		sendError(w, "Failed to create response")
		return
	}

	// Respond with the MessagePack response
	w.Header().Set("Content-Type", "application/msgpack")
	w.WriteHeader(http.StatusOK)
	w.Write(responseBytes)
}

func md5Sum(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
