package main

// matchmaker.go
import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	stdlog "log"
	"net/http"
	"os"
	"sync"

	log "github.com/go-kit/kit/log"
)

type (
	UserRequest struct {
		Id int `json:"id"`
	}

	ErrorResponse struct {
		Error string `json:"error"`
	}

	SessionResponse struct {
		WaitingForGame *[]*UserInfo
		AllSessions    *map[string]*GameSession
	}

	UserInfo struct {
		Id          int
		Status      int
		SessionUUID string
	}

	GameSession struct {
		Users []*UserInfo
	}

	MMConfig struct {
		Port             int `json:"port"`
		MaxUsersPerMM    int `json:"max_users_per_mm"`
		MaxUsersPerParty int `json:"max_users_per_party"`
	}
)

var AllSessions map[string]*GameSession

var OnlineUsers map[int]*UserInfo

var WaitingForGame []*UserInfo

var Config MMConfig

var UserMutex = &sync.Mutex{}

const (
	ConfigFile string = "config.json"
)

func readConfig(configFile string) (MMConfig, error) {
	var jsonObj MMConfig
	data, err := ioutil.ReadFile(configFile)

	if err != nil {
		return jsonObj, errors.New("Unable to open %s!")
	} else {
		jsonErr := json.Unmarshal(data, &jsonObj)

		if jsonErr != nil {
			return jsonObj, jsonErr
		}
	}
	return jsonObj, nil
}

func decodeUserRequest(req *http.Request) (UserRequest, error) {
	var userReq UserRequest
	if err := json.NewDecoder(req.Body).Decode(&userReq); err != nil {
		return userReq, err
	}
	return userReq, nil
}

func RespondSuccess(w http.ResponseWriter, response interface{}) error {
	return encodeJSONResponse(w, response, http.StatusOK)
}

func RespondError(w http.ResponseWriter, response interface{}, code int) error {
	return encodeJSONResponse(w, response, code)
}

func encodeJSONResponse(w http.ResponseWriter, response interface{}, code int) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(response)
}

func main() {
	var err error
	Config, err = readConfig(ConfigFile)

	if err != nil {
		print(err.Error())
		os.Exit(1)
	}

	// Init global variables
	OnlineUsers = make(map[int]*UserInfo)
	AllSessions = make(map[string]*GameSession)
	WaitingForGame = make([]*UserInfo, 0)

	// Bind hendlers
	router := http.NewServeMux()
	router.HandleFunc("/join", Join)
	router.HandleFunc("/leave", Leave)
	router.HandleFunc("/session", Session)

	// Init logger
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	stdlog.SetOutput(log.NewStdlibAdapter(logger))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "loc", log.DefaultCaller)
	loggedRouter := LoggingMiddleware(logger)

	// Start server
	logger.Log("Port", Config.Port)
	logger.Log("MaxUsersPerMM", Config.MaxUsersPerMM)
	logger.Log("MaxUsersPerParty", Config.MaxUsersPerParty)
	http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), loggedRouter(router))
}
