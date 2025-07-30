package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	initServer()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initServer() {
	http.HandleFunc("/do", Do)
}

const response = `{
  "data": {
    "hero": {
      "name": "R2-D2",
      "friends": [{"name": "Luke Skywalker"}, {"name": "Han Solo"},{"name": "Leia Organa"}]
    }
  }
}`

func Do(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	if err := BackendPost(); err != nil {
		log.Print(err)
		_, _ = w.Write([]byte(`{"status": "error"}`))
		return
	}

	_, _ = w.Write([]byte(response))
}

func BackendPost() error {
	body := `query HeroNameAndFriends {
      hero {
        name
        friends {
          name
        }
      }
    }`
	url := fmt.Sprintf("http://%s/process", os.Getenv("GONKEX_MOCK_BACKEND"))
	res, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("backend response status code %d", res.StatusCode)
	}

	return res.Body.Close()
}
