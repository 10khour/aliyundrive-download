package pkg

import (
	"io"
	"log"
	"net/http"
)

func printHttpBody(response *http.Response) {
	defer response.Body.Close()
	buf, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%s", string(buf))
}
