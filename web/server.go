package web

import (
	"github.com/gorilla/mux"
	"majo-tech.com/share/storage"
	"majo-tech.com/share/web/handlers"
	"majo-tech.com/share/web/templates"
	"net/http"
	"time"
)

var (
	textHtmlRegex          = "(text\\/html)|(\\*\\/\\*)"
	multipartFormDataRegex = "^multipart\\/form-data;\\ ?boundary="
)

type Server struct {
	Storage          storage.Storage
	MaxFileSizeBytes int64 // Max size for a single file upload.
	DiskSpaceBytes   int64 // Max size that the sum of all files are allowed to allocate on disk.
	Host             string
}

func (s *Server) Start() error {
	router := mux.NewRouter()

	assets, err := templates.LoadAssets()
	if err != nil {
		return err
	}

	templates, err := templates.LoadTemplates()
	if err != nil {
		return err
	}

	timeoutErrorTemplate, err := templates.TemplateTimeoutError()
	if err != nil {
		return err
	}

	router.
		Methods("GET").
		Path("/shares/{key}").
		HeadersRegexp("Accept", textHtmlRegex).
		Handler(&handlers.DownloadPageHandler{
			Templates: templates,
			Storage:   s.Storage,
		})

	router.
		Methods("POST").
		Path("/shares/{key}").
		Handler(http.TimeoutHandler(&handlers.DownloadHandler{
			Storage:   s.Storage,
			Templates: templates,
		}, time.Second*120, timeoutErrorTemplate))

	router.
		Methods("POST").
		Path("/shares").
		HeadersRegexp(
			"Content-Type", multipartFormDataRegex,
			"Accept", textHtmlRegex,
		).
		Handler(http.TimeoutHandler(&handlers.UploadHandler{
			Storage:          s.Storage,
			Templates:        templates,
			MaxFileSizeBytes: s.MaxFileSizeBytes,
			DiskSpaceBytes:   s.DiskSpaceBytes,
			Host:             s.Host,
		}, time.Second*120, timeoutErrorTemplate))

	router.
		Methods("GET").
		Path("/").
		HeadersRegexp("Accept", textHtmlRegex).
		Handler(&handlers.UploadPageHandler{
			Storage:          s.Storage,
			Templates:        templates,
			MaxFileSizeBytes: s.MaxFileSizeBytes,
		})

	router.
		Methods("GET").
		PathPrefix("/assets").
		Handler(http.StripPrefix("/assets", http.FileServer(http.FS(assets))))

	// TODO add cors header handling

	httpServer := http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8080",
		WriteTimeout: time.Second * 125,
		ReadTimeout:  time.Second * 125,
	}

	return httpServer.ListenAndServe()
}
