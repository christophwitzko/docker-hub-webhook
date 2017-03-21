package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"text/template"
	"time"
)

const okResponse = "{\"ok\":true}"
const jsonContentType = "application/json; charset=utf-8"

var validName = regexp.MustCompile("^[a-zA-Z0-9-_/]+$")
var callbackClient = &http.Client{Timeout: 10 * time.Second}
var defaultTag *regexp.Regexp
var defaultToken string
var defaultParams *template.Template
var deploysh = template.Must(template.ParseFiles("deploy.sh"))
var defaultVhost string = os.Getenv("DEFAULT_VHOST")

func init() {
	if tag := os.Getenv("DEFAULT_TAG"); tag != "" {
		defaultTag = regexp.MustCompile(tag)
	} else {
		defaultTag = regexp.MustCompile("^latest$")
	}
	if token := os.Getenv("DEFAULT_TOKEN"); token != "" {
		defaultToken = token
	} else {
		randomBytes := make([]byte, 12)
		rand.Read(randomBytes)
		defaultToken = hex.EncodeToString(randomBytes)
	}
	defaultParams = template.Must(template.New("params").Parse(os.Getenv("DEFAULT_PARAMS")))
}

type PushData struct {
	Tag    string `json:"tag"`
	Pusher string `json:"pusher"`
}

type Repository struct {
	Name     string `json:"name"`
	RepoName string `json:"repo_name"`
}

type Webhook struct {
	PushData    PushData   `json:"push_data"`
	CallbackURL string     `json:"callback_url"`
	Repository  Repository `json:"repository"`
}

type Callback struct {
	State       string `json:"state"`
	Context     string `json:"context"`
	Description string `json:"description"`
}

type TemplateData struct {
	Vhost    string
	RepoName string
	Name     string
	Tag      string
	Params   string
}

func writeError(w http.ResponseWriter, err string, code int) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": err})
}

func sendCallback(w http.ResponseWriter, url string, success bool, description string) {
	body := Callback{
		State:       "failure",
		Context:     "Webhook deploy server",
		Description: description,
	}
	if len(body.Description) > 255 {
		body.Description = body.Description[0:255]
	}
	if success {
		body.State = "success"
	}
	buff := new(bytes.Buffer)
	json.NewEncoder(buff).Encode(body)
	res, err := callbackClient.Post(url, jsonContentType, buff)
	if err != nil || res.StatusCode != 200 {
		log.Printf("invalid callback: %d", res.StatusCode)
		writeError(w, "invalid callback", http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, okResponse)
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s - %s %s", r.RemoteAddr, r.Method, r.URL.EscapedPath())
	w.Header().Set("Content-Type", jsonContentType)
	if r.Method != "POST" {
		writeError(w, "invalid method", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.EscapedPath() != "/"+defaultToken {
		writeError(w, "invalid token", http.StatusUnauthorized)
		return
	}

	var payload Webhook
	json.NewDecoder(r.Body).Decode(&payload)
	r.Body.Close()

	if payload.CallbackURL == "" ||
		!validName.MatchString(payload.Repository.Name) ||
		!validName.MatchString(payload.Repository.RepoName) {
		writeError(w, "invalid input", http.StatusBadRequest)
		return
	}
	if !defaultTag.MatchString(payload.PushData.Tag) {
		log.Printf("skipping tag: %s", payload.PushData.Tag)
		sendCallback(w, payload.CallbackURL, true, "skipped tag")
		return
	}

	info := TemplateData{
		Vhost:    defaultVhost,
		RepoName: payload.Repository.RepoName,
		Name:     payload.Repository.Name,
		Tag:      payload.PushData.Tag,
	}
	buff := new(bytes.Buffer)
	defaultParams.Execute(buff, info)
	info.Params = buff.String()
	buff.Reset()
	deploysh.Execute(buff, info)

	cmd := exec.Command("bash", "-c", buff.String())
	stdouterr, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("script error: %s", err.Error())
		sendCallback(w, payload.CallbackURL, false, "script error: "+err.Error())
		return
	}
	sendCallback(w, payload.CallbackURL, true, "successfully deployed image:\n"+string(stdouterr))
}

func main() {
	log.Printf("default tag: %s", defaultTag)
	log.Printf("default token: %s", defaultToken)
	buff := new(bytes.Buffer)
	defaultParams.Execute(buff, TemplateData{RepoName: "<RepoName>", Name: "<Name>"})
	log.Printf("default params: %s", buff.String())
	log.Printf("default vhost: %s", defaultVhost)
	log.Println("starting server on port 5000...")
	log.Fatal(http.ListenAndServe(":5000", http.HandlerFunc(handler)))
}
