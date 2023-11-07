
# JoinChunks Function

This Go package provides a function named `JoinChunks` that combines multiple data chunks stored in a Google Cloud Storage bucket into a single file and generates a corresponding metadata (`.meta`) file. The function uses the MessagePack format for request and response data.

## Execute locally

```bash
FUNCTION_TARGET=JoinChunks LOCAL_ONLY=true go run cmd/main.go 
```


## Prerequisites

Before using this package, make sure you have the following prerequisites in place:

1. **Google Cloud Storage**: You should have a Google Cloud Storage bucket where the data chunks are stored. Replace `"your-bucket-name"` with the actual bucket name in the code.

2. **Google Cloud Storage Client Credentials**: Ensure that your application has the necessary permissions to access the Google Cloud Storage bucket.

3. **MessagePack Package**: The code relies on the `github.com/vmihailenco/msgpack` package to work with MessagePack data.

4. **Functions Framework**: This package is intended to work with the Google Cloud Functions Framework for Go. You should have the framework set up for deploying and running the function.

## Function Overview

The `JoinChunks` function performs the following tasks:

1. Parses the request data, which includes a hash and the number of chunks to join.

2. Initializes the Google Cloud Storage client.

3. Creates a channel to collect chunk data concurrently.

4. Reads each chunk from the bucket in parallel and sends the chunk data to the channel.

5. Waits for all goroutines to finish and closes the channel.

6. Collects the chunk data from the channel and joins it into a single content.

7. Calculates the MD5 hash of the joined content.

8. Calculates the size of the joined content and defines the MIME type (e.g., "application/octet-stream").

9. Writes the joined content to a new file in the Google Cloud Storage bucket.

10. Creates a `.meta` file with the MD5 hash, size, and MIME type.

11. Constructs a MessagePack response containing the hash.

12. Sends the MessagePack response with appropriate headers.

## Usage

You can trigger the `JoinChunks` function by making an HTTP request to it with a MessagePack-encoded request body. The function will join the specified data chunks and provide a MessagePack-encoded response with the resulting hash.

## Error Handling

The function handles errors and sends MessagePack-encoded error responses when necessary.

## Example

Here is an example of how to use the `JoinChunks` function:

```go
import (
    "github.com/GoogleCloudPlatform/functions-framework-go/functions"
    // Other necessary imports
)

func main() {
    functions.StartHTTPServer("JoinChunks", joinChunks)
}
