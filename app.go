package main
import (
	"fmt"
  "encoding/json"
	"strconv"
	"net/http"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type IdResponse struct {
	Status string
	Id string
}

var db *sql.DB

const (
    dbName = "simple-notifications"
    dbPass = "1234"
    dbHost = "mysql"
    dbPort = "3306"
)

func main() {
	dbSource := fmt.Sprintf("root:%s@tcp(%s:%s)/%s?charset=utf8",
	 	dbPass,
		dbHost,
		dbPort,
		dbName,
	)

	var err error
	db, err = sql.Open("mysql", dbSource)
	if err != nil {
		fmt.Printf("error opening mysql connection")
    panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Logger,
		middleware.DefaultCompress,
	)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	r.Put("/client/{token}", func(w http.ResponseWriter, r *http.Request) {
		clientToken := chi.URLParam(r, "token")
		fmt.Printf("PUT client. token: %s\n", clientToken)
		stmtInsertClient, _ := db.Prepare("INSERT INTO clients (token) VALUES (?)")
		result, err := stmtInsertClient.Exec(clientToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
			return;
		}
		clientId, err := result.LastInsertId()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
			return;
		}
		response := IdResponse{"ok", strconv.FormatInt(clientId, 10)}
		// w.Write([]byte(strconv.FormatInt(clientId, 10)))
		js, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
			return;
		}
		w.Write(js)
	})
	http.ListenAndServe(":8080", r)
}
