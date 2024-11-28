package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"

	"golang.org/x/crypto/bcrypt"
)

var (
	tpl *template.Template
	db  *sql.DB
)

var store = sessions.NewCookieStore([]byte("secret-key"))

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "login.html", nil)
}

func loginAuthHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal("Error parsing")
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	session, _ := store.Get(r, "session-name")

	session.Values["username"] = username
	session.Save(r, w)
	fmt.Println("Username: ", username, "\nPassword: ", password)

	var hash string
	stmt := "SELECT Hash FROM `testdb-go`.bcrypt WHERE Username = ?;"
	row := db.QueryRow(stmt, username)
	err := row.Scan(&hash)
	fmt.Println("Hash from db: ", hash)
	if err != nil {
		fmt.Println("Error selecting Hash in db by username")
		tpl.ExecuteTemplate(w, "login.html", "Check username or password")
		return
	}
	// func CompareHashAndPassword(hashedPassword, password []byte) error
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	// return nil on success
	if err == nil {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	fmt.Println("Incorrect password")
	tpl.ExecuteTemplate(w, "login.html", "Check username and password")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	tpl.ExecuteTemplate(w, "register.html", nil)
}

func registerAuthHandler(w http.ResponseWriter, r *http.Request) {

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Lấy session
	session, _ := store.Get(r, "session-name")

	// Lấy username từ session
	username, ok := session.Values["username"].(string)
	if !ok || username == "" {
		// Nếu chưa đăng nhập, chuyển hướng về trang login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Truyền username vào template
	data := map[string]interface{}{
		"username": username,
	}
	tpl.ExecuteTemplate(w, "home.html", data)
}

func main() {
	tpl, _ = tpl.ParseGlob("templates/*.html")
	var err error
	db, err = sql.Open("mysql", "root:123456@tcp(localhost:3306)/testdb-go")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	http.HandleFunc("/home", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/loginauth", loginAuthHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/registerauth", registerAuthHandler)

	http.Handle("/web/assets", http.StripPrefix("/web/assets/", http.FileServer(http.Dir("./web/assets"))))

	http.ListenAndServe(":8080", nil)
}
