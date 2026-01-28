# SecureStorage-Go

A production-ready REST API for secure file storage with AES-256-GCM encryption at rest. Built with Go using only the standard library.

## Features

- **File Upload & Encryption**: Upload files via REST API, automatically encrypted with AES-256-GCM
- **Secure File Retrieval**: Download and decrypt files on-the-fly
- **Path Traversal Protection**: Validates all filenames to prevent directory traversal attacks
- **Container Ready**: Multi-stage Docker build running as non-root user
- **No External Dependencies**: Uses only Go standard library

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Docker (optional, for containerized deployment)
- Make (optional, for using Makefile commands)

### Running Locally

1. Clone the repository:
```bash
git clone https://github.com/your-username/secure-storage-go.git
cd secure-storage-go
```

2. Set the encryption key (must be exactly 32 characters for AES-256):
```bash
export STORAGE_KEY="your-32-character-secret-key!!"
```

3. Build and run:
```bash
make run
```

Or without Make:
```bash
go build -o bin/server ./cmd/server
STORAGE_KEY="your-32-character-secret-key!!" ./bin/server
```

The server starts on port 8080 by default.

### Using Docker

1. Build the Docker image:
```bash
make docker-build
```

2. Run the container:
```bash
docker run -d \
  --name secure-storage \
  -p 8080:8080 \
  -e STORAGE_KEY="your-32-character-secret-key!!" \
  -v $(pwd)/data:/app/data \
  secure-storage-go
```

## API Endpoints

### Upload File

```bash
POST /upload
Content-Type: multipart/form-data
```

Upload a file to be encrypted and stored.

**Example:**
```bash
curl -X POST -F "file=@myfile.pdf" http://localhost:8080/upload
```

**Response:**
```json
{
  "success": true,
  "message": "file uploaded and encrypted successfully"
}
```

### Download File

```bash
GET /download/{filename}
```

Retrieve and decrypt a previously uploaded file.

**Example:**
```bash
curl http://localhost:8080/download/myfile.pdf --output myfile.pdf
```

### Health Check

```bash
GET /health
```

Check if the service is running.

**Example:**
```bash
curl http://localhost:8080/health
```

**Response:**
```json
{
  "success": true,
  "message": "healthy"
}
```

## Security Features

### AES-256-GCM Encryption

All files are encrypted using AES-256-GCM (Galois/Counter Mode), which provides:
- **Confidentiality**: Files are encrypted with a 256-bit key
- **Integrity**: GCM provides authenticated encryption, detecting any tampering
- **Unique Nonces**: Each file uses a cryptographically random 12-byte nonce

### Path Traversal Prevention

Filenames are validated against:
- Directory traversal sequences (`..`)
- Path separators (`/`, `\`)
- Invalid characters (only alphanumeric, `-`, `_`, `.` allowed)

### Container Security

The Docker container:
- Uses a multi-stage build to minimize image size
- Runs as a non-root user (`appuser`)
- Has health checks configured
- Uses Alpine Linux for minimal attack surface

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `STORAGE_KEY` | Yes | - | 32-character AES-256 encryption key |
| `PORT` | No | 8080 | Server port |

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── api/
│   │   └── handlers.go      # HTTP handlers
│   └── storage/
│       └── storage.go       # Encryption & file storage logic
├── data/                    # Encrypted files (created at runtime)
├── Dockerfile               # Multi-stage Docker build
├── Makefile                 # Build commands
├── go.mod                   # Go module definition
└── README.md               # This file
```

## Development

### Running Tests

```bash
make test
```

### Building

```bash
make build
```

The binary will be created at `bin/server`.

### Cleaning Up

```bash
make clean
```

## Production Considerations

1. **Key Management**: Never hardcode the encryption key. Use a secrets manager like HashiCorp Vault or cloud provider secret services.

2. **HTTPS**: Deploy behind a reverse proxy (nginx, Traefik) with TLS termination.

3. **Rate Limiting**: Add rate limiting for upload endpoints to prevent abuse.

4. **Logging**: In production, consider structured logging and log aggregation.

5. **Backup**: Implement backup strategies for the `data/` directory.

## License

MIT License - feel free to use this for your portfolio or projects.

## Author

Built as a demonstration of secure file storage patterns and DevSecOps practices.
