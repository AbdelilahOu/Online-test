# Wikipedia Reverse Proxy assignment

A Go-based reverse proxy server that intercepts Wikipedia requests and modifies response links.

## Project Description

This project implements a reverse proxy server in Go that:

1. Receives incoming HTTP requests
2. Forwards them to Wikipedia
3. Modifies the response by replacing Wikipedia links with m-wikipedia.org links
4. Maintains all other HTML structure and HTTP headers
5. Returns the modified response to the client

## Installation

1. Clone this repository
2. Navigate to the project directory
3. Run the Go server:

```bash
make run
```

or:

```bash
go run main.go
```

## Usage

Once the server is running, it will listen for incoming HTTP requests. All requests will be:

1. Forwarded to Wikipedia
2. Have their response content modified
3. Returned to the client

Example transformation:

```html
Original: <a href="https://wikipedia.org/wiki/OpenAI">OpenAI</a> Modified:
<a href="https://m-wikipedia.org/wiki/OpenAI">OpenAI</a>
```

## Implementation Details

The server implements the following key components:

1. HTTP server setup
2. Request forwarding
3. Response modification
4. Header management
5. Error handling

## Success Criteria

The implementation must:

- [x] Correctly redirect requests
- [x] Modify links as specified
- [x] Maintain HTTP headers
- [x] Handle errors appropriately
- [x] Provide clear, documented code

## Testing

To test the server:

1. Start the server
2. Send requests to the local endpoint
3. Verify that responses contain modified Wikipedia links
4. Check that HTTP headers are properly maintained
