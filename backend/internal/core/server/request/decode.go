// Package request decodes shared HTTP request shapes.
package request

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func DecodeJSON(reader *http.Request, destination any) error {
	return json.NewDecoder(reader.Body).Decode(destination)
}

func DecodeStrictJSON(reader *http.Request, destination any) error {
	decoder := json.NewDecoder(reader.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}
