package server

import (
	"embed"
	"github.com/gorilla/mux"
	"majo-tech.com/share/storage"
	"majo-tech.com/share/templates"
	"net/http"
	"time"
)

var (
	//go:embed assets
	assets embed.FS

	textHtmlRegex          = "(text\\/html)|(\\*\\/\\*)"
	multipartFormDataRegex = "^multipart\\/form-data;\\ ?boundary="
)

type Server struct {
	Storage          storage.Storage
	Templates        templates.Templates
	MaxFileSizeBytes int64 // Max size for a single file upload.
	DiskSpaceBytes   int64 // Max size that the sum of all files are allowed to allocate on disk.
	Host             string
}

func (s *Server) Start() error {
	router := mux.NewRouter()

	timeoutErrorTemplate, err := TimeoutErrorTemplate(s.Templates)
	if err != nil {
		return err
	}

	router.
		Methods("GET").
		Path("/shares/{key}").
		HeadersRegexp("Accept", textHtmlRegex).
		Handler(&DownloadPageHandler{
			Templates: s.Templates,
			Storage:   s.Storage,
		})

	router.
		Methods("POST").
		Path("/shares/{key}").
		Handler(http.TimeoutHandler(&DownloadHandler{
			Storage:   s.Storage,
			Templates: s.Templates,
		}, time.Second*120, timeoutErrorTemplate))

	router.
		Methods("POST").
		Path("/shares").
		HeadersRegexp(
			"Content-Type", multipartFormDataRegex,
			"Accept", textHtmlRegex,
		).
		Handler(http.TimeoutHandler(&UploadHandler{
			Storage:          s.Storage,
			Templates:        s.Templates,
			MaxFileSizeBytes: s.MaxFileSizeBytes,
			DiskSpaceBytes:   s.DiskSpaceBytes,
			Host:             s.Host,
		}, time.Second*120, timeoutErrorTemplate))

	router.
		Methods("GET").
		Path("/").
		HeadersRegexp("Accept", textHtmlRegex).
		Handler(&UploadPageHandler{
			Storage:          s.Storage,
			Templates:        s.Templates,
			MaxFileSizeBytes: s.MaxFileSizeBytes,
		})

	router.
		Methods("GET").
		PathPrefix("/assets").
		Handler(http.FileServer(http.FS(assets)))

	httpServer := http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: time.Second * 125,
		ReadTimeout:  time.Second * 125,
	}

	return httpServer.ListenAndServe()
}
