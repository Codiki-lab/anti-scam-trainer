package response

import "net/http"

func Error(writer http.ResponseWriter, message string, status int) {
	http.Error(writer, message, status)
}
