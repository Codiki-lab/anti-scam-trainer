package response

import "net/http"

type Writer struct {
	http.ResponseWriter
	statusCode int
}

func NewWriter(writer http.ResponseWriter) *Writer {
	return &Writer{ResponseWriter: writer}
}

func (writer *Writer) WriteHeader(statusCode int) {
	writer.statusCode = statusCode
	writer.ResponseWriter.WriteHeader(statusCode)
}

func (writer *Writer) Write(body []byte) (int, error) {
	if writer.statusCode == 0 {
		writer.WriteHeader(http.StatusOK)
	}
	return writer.ResponseWriter.Write(body)
}

func (writer *Writer) Status() int {
	if writer.statusCode == 0 {
		return http.StatusOK
	}
	return writer.statusCode
}
