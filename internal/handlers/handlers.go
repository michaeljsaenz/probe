package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/michaeljsaenz/probe/internal/network"
	"github.com/michaeljsaenz/probe/internal/types"
)

func StaticFiles() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}

func BaseUrl(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func SubmitUrl(w http.ResponseWriter, r *http.Request) {
	var err error
	var serverHeader, requestedUrlHeader, requestHostnameHeader string

	url := r.PostFormValue("url")

	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")

	if url != "" {
		serverHeader, requestedUrlHeader, requestHostnameHeader, err = network.GetUrl(url)
		if err != nil {
			log.Printf("error: %v", err)
		}
	}

	// update shared context with the requestedUrlHeader
	types.UpdateSharedContext(requestedUrlHeader, requestHostnameHeader)

	application := types.NewApplication(serverHeader, requestedUrlHeader, "", []string{}, err)

	if err == nil && url != "" {
		application.HttpRequestedUrl = "HTTP Requested URL: " + requestedUrlHeader
		application.HttpServerHeader = "HTTP Server header: " + serverHeader
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	tmpl.ExecuteTemplate(w, "submit", application)

}

func ButtonCertificates(w http.ResponseWriter, r *http.Request) {
	types.ContextLock.Lock()
	defer types.ContextLock.Unlock()

	var err error
	var certificates string

	// retrieve value from shared context
	customValues, ok := types.SharedContext.Value(types.ContextKey).(types.CustomContextValues)
	if ok && customValues.URL != "" {
		certificates, err = network.GetCertificates(customValues.URL)
		if err != nil {
			log.Printf("error: %v", err)
		}
	}

	application := types.NewApplication("", customValues.URL, certificates, []string{}, err)

	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	tmpl.ExecuteTemplate(w, "certificates", application)
}

func ButtonDNS(w http.ResponseWriter, r *http.Request) {
	types.ContextLock.Lock()
	defer types.ContextLock.Unlock()

	var err error
	var ips []string

	// retrieve value from shared context
	customValues, ok := types.SharedContext.Value(types.ContextKey).(types.CustomContextValues)
	if ok && customValues.Hostname != "" {
		ips, err = network.GetDNS(customValues.Hostname)
		if err != nil {
			log.Printf("error: %v", err)
		}
	}

	application := types.NewApplication("", "", "", ips, err)

	tmpl := template.Must(template.ParseFiles("templates/index.html"))

	tmpl.ExecuteTemplate(w, "dns-lookup", application)
}
