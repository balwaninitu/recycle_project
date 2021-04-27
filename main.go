package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string
	Password []byte
}

var (
	TPL         *template.Template
	mapUsers    = map[string]User{}
	mapSessions = map[string]string{}
	DB          *sql.DB
	err         error
)

func init() {
	TPL = template.Must(template.ParseGlob("templates/*.gohtml"))
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s",
		"user", "password", "127.0.0.1:3306", "recycle_db")

	DB, err = sql.Open("mysql", dataSourceName)

	if err != nil {
		panic(err.Error())
	}
	//defer Client.Close()
	if err = DB.Ping(); err != nil {
		panic(err)
	}

	log.Println(" Connected to Database ")

}

func main() {

	r := mux.NewRouter()
	//Routers for user signup/login
	r.HandleFunc("/", Index)
	r.HandleFunc("/signup", Signup)

	http.ListenAndServe(":5221", r)

}
func Index(w http.ResponseWriter, r *http.Request) {
	myUser := GetUser(w, r)
	TPL.ExecuteTemplate(w, "index.gohtml", myUser)
}

func GetUser(w http.ResponseWriter, r *http.Request) User {
	// get current session cookie
	myCookie, err := r.Cookie("myCookie")
	if err != nil {
		id, _ := uuid.NewV4()
		myCookie = &http.Cookie{
			Name:  "myCookie",
			Value: id.String(),
		}

	}
	http.SetCookie(w, myCookie)

	// map user's cookie if exist already
	var myUser User
	if username, ok := mapSessions[myCookie.Value]; ok {
		myUser = mapUsers[username]
	}

	return myUser
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if AlreadyLoggedIn(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var myUser User
	// process form submission
	if r.Method == http.MethodPost {
		// get form values
		username := r.FormValue("username")
		password := r.FormValue("password")
		if username == "" || password == "" {
			//todo security features
		}

		results, err := DB.Query("SELECT Username, Password from users")
		if err != nil {
			panic(err)
		}
		checkUsers := map[string]User{}
		var exixtingUser User
		for results.Next() {
			err = results.Scan(&exixtingUser.Username, &exixtingUser.Password)
			if err != nil {
				panic(err)
			}
			checkUsers[exixtingUser.Username] = exixtingUser
		}
		bPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		myUser = User{username, bPassword}
		mapUsers[username] = myUser
		query := fmt.Sprintf("Insert INTO users VALUES ('%s','%s')", myUser.Username, myUser.Password)
		_, err = DB.Query(query)
		if err != nil {
			panic(err)
		}
		id, _ := uuid.NewV4()
		myCookie := &http.Cookie{
			Name:    "myCookie",
			Value:   id.String(),
			Expires: time.Now().Add(5 * time.Minute),
		}
		http.SetCookie(w, myCookie)
		mapSessions[myCookie.Value] = username
	}
	TPL.ExecuteTemplate(w, "signup.gohtml", nil)
}
func AlreadyLoggedIn(r *http.Request) bool {
	myCookie, err := r.Cookie("myCookie")
	if err != nil {
		return false
	}
	username := mapSessions[myCookie.Value]
	_, ok := mapUsers[username]
	return ok
}
