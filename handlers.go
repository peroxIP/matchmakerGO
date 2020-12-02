package main

// handlers.go
import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// User joins matchmaking
func Join(w http.ResponseWriter, req *http.Request) {

	var u UserRequest
	err := json.NewDecoder(req.Body).Decode(&u)

	if err != nil {
		RespondError(w, ErrorResponse{Error: err.Error()}, http.StatusBadRequest)
		return
	}

	if len(OnlineUsers) == Config.MaxUsersPerMM {
		RespondError(w, ErrorResponse{Error: err.Error()}, http.StatusLocked)
		return
	}
	UserMutex.Lock()

	_, found := OnlineUsers[u.Id]
	if found {
		RespondError(w, ErrorResponse{Error: "User already in queue"}, http.StatusAlreadyReported)
	} else {
		var userInfo UserInfo

		userInfo.Status = 1
		userInfo.Id = u.Id

		// Mark user as online and queue them for game session
		OnlineUsers[u.Id] = &userInfo
		WaitingForGame = append(WaitingForGame, &userInfo)

		tryFormSession()
	}

	UserMutex.Unlock()
}

// User leaves the matchmaking
func Leave(w http.ResponseWriter, req *http.Request) {
	userReq, err := decodeUserRequest(req)
	if err != nil {
		RespondError(w, ErrorResponse{Error: err.Error()}, http.StatusBadRequest)
		return
	}
	UserMutex.Lock()

	userInfo, found := OnlineUsers[userReq.Id]
	if found {
		if userInfo.Status == 1 {
			// User is waiting for a game
			findAndRemove(&WaitingForGame, userInfo)
		} else if userInfo.Status == 2 {
			// User is in session
			session := AllSessions[userInfo.SessionUUID]

			// Remove user from session
			findAndRemove(&session.Users, userInfo)

			// Reset the remaining users status for waiting
			for i := 0; i < len(session.Users); i++ {
				temp := session.Users[i]
				temp.SessionUUID = ""
				temp.Status = 1
			}
			// Return remainging users to waiting for session
			WaitingForGame = append(WaitingForGame, session.Users...)

			// Remove disbanded session
			delete(AllSessions, userInfo.SessionUUID)
		}

		// User left matchmaking
		delete(OnlineUsers, userInfo.Id)

		tryFormSession()
	} else {
		RespondError(w, ErrorResponse{Error: "User not found"}, http.StatusNotFound)
	}

	UserMutex.Unlock()
}

// returns all session of games and waiting users
func Session(w http.ResponseWriter, req *http.Request) {
	var response SessionResponse
	response.AllSessions = &AllSessions
	response.WaitingForGame = &WaitingForGame

	err := RespondSuccess(w, response)
	if err != nil {
		RespondError(w, ErrorResponse{Error: err.Error()}, http.StatusInternalServerError)
		return
	}
}

// Check if there are enough users to form a session
func tryFormSession() {
	if len(WaitingForGame) >= Config.MaxUsersPerParty {

		sessionUUID := uuid.New().String()

		// Pop the the users that will form a session
		sessionUserList := WaitingForGame[:Config.MaxUsersPerParty]
		WaitingForGame = WaitingForGame[Config.MaxUsersPerParty:]

		var session GameSession

		// set the status of the users
		for i := 0; i < len(sessionUserList); i++ {
			userInfo := sessionUserList[i]
			userInfo.SessionUUID = sessionUUID
			userInfo.Status = 2
		}

		session.Users = append(session.Users, sessionUserList...)

		// bind session
		AllSessions[sessionUUID] = &session
	}
}

// Find a specified user in an array
func findAndRemove(users *[]*UserInfo, find *UserInfo) {
	var i int
	var temp []*UserInfo

	temp = *users
	for i = 0; i < len(temp); i++ {
		if temp[i] == find {
			break
		}
	}
	*users = remove(*users, i)
}

// Remove user from an array
func remove(s []*UserInfo, i int) []*UserInfo {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
