package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func errJSON(w http.ResponseWriter, err error) {
	fmt.Printf("error: |%s|\n", err)
	w.WriteHeader(200)
	fmt.Fprintf(w, `{"error":"internal-server-error"}`+"\n")
}

func okJSON(w http.ResponseWriter, out interface{}) {
	b, err := json.Marshal(out)
	if err != nil {
		errJSON(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s\n", string(b))
}
