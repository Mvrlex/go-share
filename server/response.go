package server

import (
	"bytes"
	"majo-tech.com/share/templates"
	"net/http"
)

func WriteError(writer http.ResponseWriter, templates templates.Templates, status int, desc string) {
	writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(status)
	err := templates.TemplateError(writer, desc)
	if err != nil {
		http.Error(writer, desc, status)
	}
}

func WriteErrorPage(writer http.ResponseWriter, templates templates.Templates, status int, desc string) {
	writer.Header().Add("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(status)
	err := templates.TemplateError(writer, desc)
	if err != nil {
		http.Error(writer, desc, status)
	}
}

func TimeoutErrorTemplate(templates templates.Templates) (string, error) {
	var tpl bytes.Buffer
	err := templates.TemplateError(&tpl, "Your request took too long, so for security reasons, we closed the connection.")
	if err != nil {
		return "", err
	}
	return tpl.String(), nil
}
