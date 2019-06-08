package main
import (
	"fmt"
	"log"
	"os"
	"bytes"
	"io/ioutil"
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

type SubscriptionRequest struct {
	ClientId int	`json:"clientId"`
	TopicId int		`json:"topicId"`
}

type FCMNotification struct {
	Title string 	`json:"title"`
	Body string 	`json:"body"`
}

type FCMMessageJSON struct {
	To string											`json:"to"`
	Notification FCMNotification	`json:"notification"`
}

var db *sql.DB

var (
	dbName,
	dbPass,
	dbHost,
	dbPort,
	testToken,
	webhook,
	auth_header string
)

func LoadConf() {
	dbName = os.Getenv("DB__NAME") // "simple-notifications"
	dbPass = os.Getenv("DB__PASS") // "1234"
	dbHost = os.Getenv("DB__HOST") // "mysql"
	dbPort = os.Getenv("DB__PORT") // "3306"
	testToken = os.Getenv("TEST_TOKEN") // "dFD_xVCueR8:APA91bEmLFf7-7R--HO3PFsVGacKHCnJ0K2bhsdRaM7hhgRbgeZijbk1jysjqylQU36K58FFQeooqIub3a180JeTWbfPK37YoVEW6M1cM5TfgH1P1kd26eYnghh0m437uJ5CL3usKhzb"
	webhook = os.Getenv("WEBHOOK") // "https://fcm.googleapis.com/fcm/send"
	auth_header = os.Getenv("AUTH_HEADER") // "key=AAAA_0UuTtQ:APA91bEaPoxKJeT00DAgRpQXC4dfJaNqsRUkxNj6UMe-IUh1CfcQsJ3AZMTceT9HX2u06mznkr08-Ee_mpV9rmJKa4JSWmvjszrGJPf5UYstpW3BvseP9XIFR9VKqUpdASJIE23xb1nd"
}

func sendMessage(title string, msg string, to string) {
	notification := FCMNotification{title, msg}
	messageJson := FCMMessageJSON{to, notification}
	client := &http.Client{}

	js, err := json.Marshal(messageJson)
	if err != nil {
		log.Println("Error marshaling json for FCMMessageJSON")
		return;
	}
	debug_js := string(js)
	log.Println(debug_js)
	req, err := http.NewRequest("POST", webhook, bytes.NewBuffer(js))
	if err != nil {
		log.Println("Error creating http request to webhook")
		log.Println(err)
		return;
	}
	req.Header.Add("Authorization", auth_header)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error while trying to send request to webhook", err)
		return;
	}
  defer resp.Body.Close()
}

func main() {
	log.Println("Starting...")
	LoadConf()
	dbSource := fmt.Sprintf("root:%s@tcp(%s:%s)/%s?charset=utf8",
	 	dbPass,
		dbHost,
		dbPort,
		dbName,
	)

	var err error
	db, err = sql.Open("mysql", dbSource)
	if err != nil {
		log.Printf("error opening mysql connection")
    panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	r := chi.NewRouter()
	r.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.Logger,
		middleware.DefaultCompress,
	)

	r.Put("/client/{token}", func(w http.ResponseWriter, r *http.Request) {
		clientToken := chi.URLParam(r, "token")
		log.Printf("PUT client. token: %s\n", clientToken)
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

	r.Put("/topic/{name}", func(w http.ResponseWriter, r *http.Request) {
		topicName := chi.URLParam(r, "name")
		log.Printf("PUT topic. name: %s\n", topicName)
		stmtInsertTopic, _ := db.Prepare("INSERT INTO topics (name) VALUES (?)")
		result, err := stmtInsertTopic.Exec(topicName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
			return;
		}
		topicId, err := result.LastInsertId()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
			return;
		}
		response := IdResponse{"ok", strconv.FormatInt(topicId, 10)}
		// w.Write([]byte(strconv.FormatInt(clientId, 10)))
		js, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
			return;
		}
		w.Write(js)
	})

	r.Post("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		var subscriptionRequest SubscriptionRequest
		b, err := ioutil.ReadAll(r.Body)
		log.Printf("raw body: %s\n", b)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Unmarshal
		err = json.Unmarshal(b, &subscriptionRequest)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		out, err := json.Marshal(subscriptionRequest)
		log.Printf("subscriptionRequest parsed to %s\n", out)
		log.Printf("POST subscribe. clientId: %d, topicId: %d\n", subscriptionRequest.ClientId, subscriptionRequest.TopicId)
		stmtInsertSubscription, _ := db.Prepare("INSERT INTO subscriptions (client_id, topic_id) VALUES (?, ?)")
		result, err := stmtInsertSubscription.Exec(
			subscriptionRequest.ClientId, subscriptionRequest.TopicId)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
			return;
		}

		subscriptionId, err := result.LastInsertId()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
			return;
		}
		response := IdResponse{"ok", strconv.FormatInt(subscriptionId, 10)}
		// w.Write([]byte(strconv.FormatInt(clientId, 10)))
		js, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 - Something bad happened!"))
			return;
		}
		w.Write(js)
	})

	r.Post("/testmessage", func(w http.ResponseWriter, r *http.Request) {
		to := testToken
		go sendMessage("test title", "test message", to)
		w.Write([]byte("OK"))
	})

	r.Post("/topic/{name}", func(w http.ResponseWriter, r *http.Request) {
		topicName := chi.URLParam(r, "name")
		log.Printf("POST topic. name: %s\n", topicName)
		w.Write([]byte("NOT IMPLEMENTED"))
	})

	log.Println("Listening on port 8080")
	http.ListenAndServe(":8080", r)
}
