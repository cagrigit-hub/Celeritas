package celeritas

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

const version = "1.0.0"

type Celeritas struct {
	AppName  string
	Debug    bool
	Version  string
	Errorlog *log.Logger
	Infolog  *log.Logger
	RootPath string
	Routes   *chi.Mux
	config   config
}

type config struct {
	port     string
	renderer string
}

func (c *Celeritas) New(rootPath string) error {
	pathConfig := initPaths{
		RootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "data", "public", "tmp", "logs", "middleware"},
	}
	err := c.Init(pathConfig)
	if err != nil {
		return err
	}

	err = c.checkDotEnv(rootPath)
	if err != nil {
		return err
	}

	// read .env
	err = godotenv.Load(fmt.Sprintf("%s/.env", rootPath))
	if err != nil {
		return err
	}

	// create loggers
	infoLog, errorLog := c.startLoggers(rootPath)
	c.Infolog = infoLog
	c.Errorlog = errorLog
	c.Debug = os.Getenv("DEBUG") == "true"
	c.RootPath = rootPath
	c.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
	}
	c.Routes = c.routes().(*chi.Mux)
	return nil
}

func (c *Celeritas) Init(p initPaths) error {
	root := p.RootPath
	for _, folder := range p.folderNames {
		err := c.CreateDirIfNotExist(filepath.Join(root, folder))
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Celeritas) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", c.config.port),
		ErrorLog:     c.Errorlog,
		Handler:      c.Routes,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	c.Infolog.Printf("Starting %s on port %s", c.AppName, c.config.port)
	err := srv.ListenAndServe()
	c.Errorlog.Fatal(err)
}

func (c *Celeritas) checkDotEnv(rootPath string) error {
	err := c.CreateDirIfNotExist(fmt.Sprintf("%s/.env", rootPath))
	if err != nil {
		return err
	}
	return nil
}

func (c *Celeritas) startLoggers(rootPath string) (*log.Logger, *log.Logger) {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)

	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLog, errorLog
}
