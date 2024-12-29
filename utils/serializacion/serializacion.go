package serialization

import (
	"encoding/json"
	"net/http"
)

// Decode el body de una solicitud HTTP en generic
// Usar cuando queremos leer un JSON que viene del body y pasarlo a una struct de Go
func DecodeHTTPBody[T any](r *http.Request, data T) error {
	return json.NewDecoder(r.Body).Decode(data)
}

// Encode la estructura en JSON y la escribe en el body de la respuesta HTTP
// Usar cuando queres responder con un JSON
func EncodeHTTPResponse[T any](w http.ResponseWriter, data T, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
