// Package request decodes shared HTTP request shapes.
package request

import (
	"encoding/json"
	"net/http"
)

func DecodeJSON(reader *http.Request, destination any) error {
	return json.NewDecoder(reader.Body).Decode(destination)
}
