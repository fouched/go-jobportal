package main

import (
	"flag"
	"fmt"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/v2"
	"github.com/fouched/go-jobportal/internal/driver"
	"github.com/fouched/go-jobportal/internal/models"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

type config struct {
	port     int
	env      string
	dsn      string
	frontend string // used to check for URL tampering during password reset
}

type application struct {
	config        config
	infoLog       *log.Logger
	errorLog      *log.Logger
	templateCache map[string]*template.Template
	version       string
	DB            models.DBModel
	Session       *scs.SessionManager
}

const version = "1.0.0"
const cssVersion = "1"

var session *scs.SessionManager

func main() {
	// register struct types that needs to be stored in the session
	//gob.Register(TransactionData{})

	var cfg config

	flag.IntVar(&cfg.port, "port", 9080, "Server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Application environment {development|production}")
	flag.StringVar(&cfg.dsn, "dsn", "jobportal:jobportal@tcp(localhost:3306)/gojobportal?parseTime=true&tls=false", "DSN")
	flag.StringVar(&cfg.frontend, "frontend", "http://localhost:9080", "url to front end")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// set up db conn
	conn, err := driver.OpenDB(cfg.dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer conn.Close()

	// set up session
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	// we use persistent storage iso cookies for session data, this allows us to
	// restart the server without users losing the login / session information
	// https://github.com/alexedwards/scs has various options available
	session.Store = mysqlstore.New(conn)

	tc := make(map[string]*template.Template)

	app := &application{
		config:        cfg,
		infoLog:       infoLog,
		errorLog:      errorLog,
		templateCache: tc,
		version:       version,
		DB:            models.DBModel{DB: conn},
		Session:       session,
	}

	err = app.serve()
	if err != nil {
		app.errorLog.Println(err)
		log.Fatalln(err)
	}
}

func (app *application) serve() error {

	//set up renderer
	//render.NewRenderer(&app)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(),
	}
	app.infoLog.Println(fmt.Sprintf("Starting HTTP server in %s mode on port %d", app.config.env, app.config.port))

	return srv.ListenAndServe()

}
