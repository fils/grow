package fileactions

import (
	"mime"
	"net/http"
	"os"
)

// MimeByType matches file extensions to mimetype
func MimeByType(e string) string {
	t := mime.TypeByExtension(e)
	if t == "" {
		t = "application/octet-stream"
	}
	return t
}

// GetFileContentType works for os.File, but what about minio
func GetFileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}
