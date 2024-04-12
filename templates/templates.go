package templates

import (
	"embed"
	"errors"
	"html/template"
	"io"
)

//go:embed templates
var templates embed.FS

var TemplateLoadError = errors.New("could not load templates")

type Templates struct {
	UploadPage   *template.Template
	Share        *template.Template
	DownloadPage *template.Template
	Error        *template.Template
}

func Load() (Templates, error) {
	uploadPageTemplate, err := template.ParseFS(templates, "templates/index.gohtml", "templates/upload.page.gohtml", "templates/file-input.component.gohtml")
	if err != nil {
		return Templates{}, errors.Join(TemplateLoadError, err)
	}
	shareTemplate, err := template.ParseFS(templates, "templates/index.gohtml", "templates/upload-success.gohtml")
	if err != nil {
		return Templates{}, errors.Join(TemplateLoadError, err)
	}
	downloadPageTemplate, err := template.ParseFS(templates, "templates/index.gohtml", "templates/download.page.gohtml")
	if err != nil {
		return Templates{}, errors.Join(TemplateLoadError, err)
	}
	errorTemplate, err := template.ParseFS(templates, "templates/index.gohtml", "templates/error.gohtml")
	if err != nil {
		return Templates{}, errors.Join(TemplateLoadError, err)
	}
	return Templates{
		UploadPage:   uploadPageTemplate,
		Share:        shareTemplate,
		DownloadPage: downloadPageTemplate,
		Error:        errorTemplate,
	}, nil
}

func (t *Templates) TemplateUpload(wr io.Writer, maxBytes int64, currBytes int64) error {
	return t.UploadPage.ExecuteTemplate(wr, "layout", struct {
		MaxFileSizeBytes     int64
		CurrentFileSizeBytes int64
	}{
		MaxFileSizeBytes:     maxBytes,
		CurrentFileSizeBytes: currBytes,
	})
}

func (t *Templates) TemplateUploadSuccess(wr io.Writer, qrCode []byte, link string) error {
	return t.Share.ExecuteTemplate(wr, "layout", struct {
		QrCode string
		Link   string
	}{
		QrCode: string(qrCode),
		Link:   link,
	})
}

func (t *Templates) TemplateDownload(wr io.Writer, key string, fileName string, fileSize string, requiresPassword bool) error {
	return t.DownloadPage.ExecuteTemplate(wr, "layout", struct {
		Key              string
		FileName         string
		FileSize         string
		RequiresPassword bool
	}{
		Key:              key,
		FileName:         fileName,
		FileSize:         fileSize,
		RequiresPassword: requiresPassword,
	})
}

func (t *Templates) TemplateError(wr io.Writer, error string) error {
	return t.Error.ExecuteTemplate(wr, "layout", struct {
		Error string
	}{
		Error: error,
	})
}
