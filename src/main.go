package main

import (

	"fmt"
	"html/template"
	"net/http"
	"encoding/json"
	"strings"
	"strconv"
	"time"

	"github.com/apexskier/httpauth"
	"golang.org/x/crypto/bcrypt"
)

type myMux map[string]map[string]func(http.ResponseWriter, *http.Request)

type myHandler struct {}

type Notification struct {
	Type string
	Info string
	Seen bool
}

var (

	backend     httpauth.SqlAuthBackend
	aaa         httpauth.Authorizer
	roles       map[string]httpauth.Role
	port        = 8009
	backenddb   = "david:david123@tcp(127.0.0.1:3306)/flagz?parseTime=true&loc=Local"

	game          map[string]*Buscaminas // player's game
	players       map[int][2]string      // game's players
	notifications map[string][]int       // player's notifications
	notif         map[int]*Notification
	privateGame   map[int]bool
	idGame        int = 0
	idNots	      int = 0
	
	mux       myMux
	filehttp  = http.NewServeMux()
	wshttp    = http.NewServeMux()	

	connPlayer map[string]*connection
)

func (m *myMux) addRoute(path string, f func(http.ResponseWriter, *http.Request), methods []string) {

	for i := range methods {
		_, ok := (*m)[methods[i]]
		if ok == false {
			(*m)[methods[i]] = make(map[string]func(http.ResponseWriter, *http.Request))
		}
		(*m)[methods[i]][path] = f
	}
}


func main() {

	var err error

	// create the backend
	backend, err = httpauth.NewSqlAuthBackend("mysql", backenddb)
	if err != nil {
		panic(err)
	}
	defer backend.Close()

	game          = make(map[string]*Buscaminas)
	players       = make(map[int][2]string)
	connPlayer    = make(map[string]*connection)
	notifications = make(map[string][]int)
	notif         = make(map[int]*Notification)
	privateGame   = make(map[int]bool)

	// create some default roles
	roles = make(map[string]httpauth.Role)
	roles["user"] = 30
	roles["admin"] = 80
	aaa, err = httpauth.NewAuthorizer(backend, []byte("cookie-encryption-key"), "user", roles)

	// create a default user
	hash, err := bcrypt.GenerateFromPassword([]byte("adminadmin"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	defaultUser := httpauth.UserData{Username: "admin", Email: "admin@localhost", Hash: hash, Role: "admin"}
	err = backend.SaveUser(defaultUser)
	if err != nil {
		panic(err)
	}


	// Handles
	mux = make(map[string]map[string]func(http.ResponseWriter, *http.Request))
	
	// addRoute(route, handleFunction, methods)
	mux.addRoute("/", handleHome, []string{"GET", "POST"})

	mux.addRoute("/login", getLogin, []string{"GET"})
	mux.addRoute("/login", postLogin, []string{"POST"})
	mux.addRoute("/register", postRegister, []string{"POST"})
	mux.addRoute("/logout", handleLogout, []string{"GET", "POST"})

	mux.addRoute("/lobby", handleLobby, []string{"GET", "POST"})
	mux.addRoute("/lobby/games", handleLobbyGames, []string{"GET", "POST"})
	mux.addRoute("/lobby/players", handleLobbyPlayers, []string{"GET", "POST"})
	mux.addRoute("/lobby/challenge", handleLobbyChallenge, []string{"GET", "POST"})

	mux.addRoute("/game", handleGame, []string{"GET", "POST"})
	mux.addRoute("/game/init", handleGameInit, []string{"GET", "POST"})
	mux.addRoute("/game/move", handleGameMove, []string{"GET", "POST"})
	mux.addRoute("/game/data", handleGameData, []string{"GET", "POST"})
	mux.addRoute("/game/joinGame", handleGameJoinGame, []string{"GET", "POST"})
	mux.addRoute("/game/exit", handleGameExit, []string{"GET", "POST"})

	mux.addRoute("/misc/notification/get", handleNotificationGet, []string{"GET", "POST"})
	mux.addRoute("/misc/notification/reject/game", handleNotificationRejectGame, []string{"GET", "POST"})
	mux.addRoute("/misc/notification/seen", handleNotificationSeen, []string{"GET", "POST"})

	hub := newHub()
   	go hub.run()
	
	filehttp.Handle("/", http.FileServer(http.Dir("../"))) // files
	wshttp.Handle("/ws", wsHandler{h: hub})  // websocket 

	fmt.Printf("Server running on port %d\n", port)

	var mh *myHandler
	go CleanLoggedUsers()
	http.ListenAndServe(fmt.Sprintf(":%d", port), mh)
}

/******************************/
/*         Functions          */
/******************************/

func sendCommand(command, username string) {
	
	thisGame := game[username]

	var rival string
	if thisGame.Players[0] == username {
		rival = thisGame.Players[1]
	} else {
		rival = thisGame.Players[0]
	}

	conn, ok := connPlayer[rival]
	if ok {
		conn.send <- []byte("\\move")
	}
}

func deletePlayerFromGame(username string) {
	
	if _, ok := game[username]; ok == false {
		return
	}

	sendCommand("\\exit", username)

	thisGame := game[username]
	gameId   := thisGame.Id

	delete(game, username)
	
	if thisGame.Players[0] == username {
		thisGame.Players[0] = ""
	} else if thisGame.Players[1] == username {
		thisGame.Players[1] = ""
	}
	players[gameId] = thisGame.Players

	if thisGame.Players[0] == "" && thisGame.Players[1] == "" {
		delete(players, gameId)
		thisGame = nil
	}
}

func CleanLoggedUsers() {
	for ;; {
		time.Sleep(3*1000*1000*10)
		players, _ := backend.Players()
		for i := range players {

			duration := time.Since(players[i].LastSeen)
			
			if duration.Seconds() > 30 {
				backend.UpdateLogged(false, players[i].Username)
			}
		}
	}
}

func createGame(username string) int {
	var new_game Buscaminas

	game[username] = &new_game
	game[username].Init(16, 16, 51, username, idGame)
	players[idGame] = [2]string{username, ""}
	idGame++

	return idGame-1
}

func addNotification(username, tipo, info string) int {

	notif[idNots] = &Notification{Type:tipo, Info:info, Seen:false}
	notifications[username] = append(notifications[username], idNots)

	idNots++
	return idNots-1
}

/******************************/
/*           Server           */
/******************************/

func (*myHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	//fmt.Println(req.URL.Path, req.Method)

	if (strings.Contains(req.URL.Path, "/ws")) {
		wshttp.ServeHTTP(rw, req)
		return;
	}

	if (strings.Contains(req.URL.Path, ".")) {
		filehttp.ServeHTTP(rw, req)
		return;
	}

	if f, ok := mux[req.Method][req.URL.Path]; ok {

		user, err := aaa.CurrentUser(rw, req)
		if err != nil && req.URL.Path != "/login" && req.URL.Path != "/register" {

			http.Redirect(rw, req, "/login", http.StatusSeeOther)
		} else {

			if (err == nil) {

				lastTime, _  := backend.GetLastSeen(user.Username)
				duration     := time.Since(lastTime)
			
				if duration.Minutes() > 5 {
					handleLogout(rw, req)
					return
				} else {
					backend.UpdateLogged(true, user.Username)
					backend.UpdateLastSeen(user.Username)
				}

				_, ok     := game[user.Username]
				splitPath := strings.Split(req.URL.Path, "/")

				if ok && splitPath[1] != "game" && (splitPath[1] == "lobby" || splitPath[1] == "login") {
					http.Redirect(rw, req, "/game", http.StatusSeeOther)
					return
				}

				if !ok && splitPath[1] == "game" && len(splitPath) < 3 {
					http.Redirect(rw, req, "/lobby", http.StatusSeeOther)
					return
				}
			}

			f(rw, req)
		}
	}
}

func handleHome(rw http.ResponseWriter, req *http.Request) {

	if err := aaa.Authorize(rw, req, true); err != nil {
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
	} else {
		http.Redirect(rw, req, "/lobby", http.StatusSeeOther)
	}
}

/******************************/
/*            Misc            */
/******************************/

func handleNotificationGet(rw http.ResponseWriter, req *http.Request) {

	user, _   := aaa.CurrentUser(rw, req)
	
	type Response struct {
		IdNot int
		Not   *Notification
	}

	var resp []Response
	for _, idNot := range notifications[user.Username] {
		resp = append(resp, Response{IdNot:idNot, Not:notif[idNot]})
	}

	respJson, _ := json.Marshal(resp)
	rw.Write(respJson)
}

func handleNotificationRejectGame(rw http.ResponseWriter, req *http.Request) {

	gameId, _ := strconv.Atoi(req.FormValue("gameId"))
	rival     := req.FormValue("rival")
	user, _   := aaa.CurrentUser(rw, req)

	delete(privateGame, gameId)	

	addNotification(rival, "reject", user.Username)
}

func handleNotificationSeen(rw http.ResponseWriter, req *http.Request) {

	notId, _ := strconv.Atoi(req.FormValue("notId"))

	not := notif[notId]
	not.Seen = true
}

/******************************/
/*          Lobby             */
/******************************/

func handleLobby(rw http.ResponseWriter, req *http.Request) {

	user, _ := aaa.CurrentUser(rw, req)
	t, err  := template.ParseFiles("../views/lobby.html")
	if err != nil {
		panic (err)
	}
	t.Execute(rw, user)

}

func handleLobbyGames(rw http.ResponseWriter, req *http.Request) {
	

	type myResp struct {
		GameId int
		Joinable int
		Players [2]string
	}

	var resp []myResp
	resp = make([]myResp, len(players))

	var i int = 0
	for gameId, usernames := range players {
		var joinable int = 0
		_, ok := privateGame[gameId]

		if (usernames[0] == "" || usernames[1] == "") && !ok {
			joinable = 1
		}

		resp[i] = myResp{GameId:gameId, Joinable:joinable, Players:usernames}
		i++
	}

	respJson, _ := json.Marshal(resp)
	rw.Write(respJson)
}

func handleLobbyPlayers(rw http.ResponseWriter, req *http.Request) {

	type myResp struct {
		Username string
	}

	var resp []myResp
	p, _ := backend.Players()
	for i := range p {
		if p[i].Logged {
			resp = append(resp, myResp{Username:p[i].Username})
		}
	}

	respJson, _ := json.Marshal(resp)
	rw.Write(respJson)
}

func handleLobbyChallenge(rw http.ResponseWriter, req *http.Request) {

	var rival string = req.FormValue("rival")
	user, _ := aaa.CurrentUser(rw, req)

	if rival == user.Username {
		return
	}

	idGame := createGame(user.Username)
	privateGame[idGame] = true

	addNotification(rival, "challenge", user.Username+"."+strconv.Itoa(idGame))
}

/******************************/
/*           Game             */
/******************************/

func handleGameExit(rw http.ResponseWriter, req *http.Request) {

	user, err := aaa.CurrentUser(rw, req)
	if (err != nil) {
		panic(err)
	}

	deletePlayerFromGame(user.Username)
	http.Redirect(rw, req, "/lobby", http.StatusSeeOther)
}

func handleGameJoinGame(rw http.ResponseWriter, req *http.Request) {

	gameId, _ := strconv.Atoi(req.FormValue("id"))

	if _, ok := players[gameId]; ok == false {
		fmt.Println("The game doesn't exist")
		http.Redirect(rw, req, "/lobby", http.StatusSeeOther)
		return
	}

	var versus string
	if players[gameId][0] != "" {
		versus = players[gameId][0]
	} else {
		versus = players[gameId][1]
	}
	thisGame  := game[versus]
	user, _   := aaa.CurrentUser(rw, req)

	if thisGame.Players[1] == "" {
		thisGame.Players[1] = user.Username
	} else if thisGame.Players[0] == "" {
		thisGame.Players[0] = user.Username
	} else {
		fmt.Println("Already has two players.");
		return
	}

	players[gameId] = thisGame.Players
	game[user.Username] = thisGame
	delete(privateGame, gameId)

	sendCommand("\\join", user.Username)
	http.Redirect(rw, req, "/game", http.StatusSeeOther)
}

func handleGameInit(rw http.ResponseWriter, req *http.Request) {

	user, err := aaa.CurrentUser(rw, req)
	if (err != nil) {
		panic(err)
	} else {

		if _, ok := game[user.Username]; ok == true {
			fmt.Println("It's playing already.")
		} else {
			createGame(user.Username)
			
			fmt.Println("Game created.")
			game[user.Username].PrintBoard()
		}

		http.Redirect(rw, req, "/game", http.StatusSeeOther)
	}
}

func handleGame(rw http.ResponseWriter, req *http.Request) {

	user, err := aaa.CurrentUser(rw, req)
	if (err != nil) {
		panic(err)
	} else {

		t, _ := template.ParseFiles("../views/game.html")	

		type Response struct {
			Host string
			Username string
		}
		resp := Response{Host: req.Host, Username: user.Username}

		t.Execute(rw, resp)
	}
	
}

func handleGameMove(rw http.ResponseWriter, req *http.Request) {

	user, err := aaa.CurrentUser(rw, req)
	if err != nil {
		panic(err)
	}

	thisGame := game[user.Username]

	req.ParseForm()

	visited     := req.Form["visited[]"]
	usedBomb, _ := strconv.ParseBool(req.Form["usedBomb"][0])
	lastX, _    := strconv.Atoi(req.Form["lastX"][0])
	lastY, _    := strconv.Atoi(req.Form["lastY"][0])

	coord := make([][2]int16, 0, thisGame.R*thisGame.C)
	for i := 0; i < len(visited); {

		sx, sy := visited[i], visited[i+1]
		ix, _  := strconv.Atoi(sx)
		iy, _  := strconv.Atoi(sy)
		coord = append(coord, [2]int16{int16(ix), int16(iy)})

		i += 2
	}

	thisGame.Move(coord, usedBomb, int16(lastX), int16(lastY))
	sendCommand("\\move", user.Username)
}

func handleGameData(rw http.ResponseWriter, req *http.Request) {

	user, err := aaa.CurrentUser(rw, req)
	if err != nil {
		panic(err)
	}

	if _, ok := game[user.Username]; ok == false {
		fmt.Println("The user has no game.")
		return
	}

	thisGame     := game[user.Username]

	type Response struct {

		Game Buscaminas
		Username string
	}

	resp := Response{Game: *thisGame, Username: user.Username}
	respJson, _ := json.Marshal(resp)

	rw.Write(respJson)
}

/******************************/
/*          Session           */
/******************************/

func getLogin(rw http.ResponseWriter, req *http.Request) {

	if _, err := aaa.CurrentUser(rw, req); err == nil {
		http.Redirect(rw, req, "/lobby", http.StatusSeeOther)
		return
	}

	messages := aaa.Messages(rw, req)
	t, err := template.ParseFiles("../views/login.html")
	if err != nil {
		panic(err)
	}

    t.Execute(rw, messages)
}

func postLogin(rw http.ResponseWriter, req *http.Request) {
	username := req.PostFormValue("username")
	password := req.PostFormValue("password")
	err := aaa.Login(rw, req, username, password, "/lobby")

	if err != nil && err.Error() == "already authenticated" {
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
	} else if err != nil {
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
	} else {
		err = backend.UpdateLastSeen(username)
		if err != nil {
			panic(err)
		}
		err = backend.UpdateLogged(true, username)
		if err != nil {
			panic(err)
		}
	}
}

func postRegister(rw http.ResponseWriter, req *http.Request) {
	var user httpauth.UserData
	user.Username = req.PostFormValue("username")
	user.Email = req.PostFormValue("email")
	password := req.PostFormValue("password")
	if err := aaa.Register(rw, req, user, password); err == nil {
		http.Redirect(rw, req, "/login?success=1", http.StatusSeeOther)
	} else {
		http.Redirect(rw, req, "/login?success=0", http.StatusSeeOther)
	}
}

func handleLogout(rw http.ResponseWriter, req *http.Request) {
	
	user, _ := aaa.CurrentUser(rw, req)
	deletePlayerFromGame(user.Username)

	if err := aaa.Logout(rw, req); err != nil {
		fmt.Println(err)
		// this shouldn't happen
		return
	}

	backend.UpdateLogged(false, user.Username)
	http.Redirect(rw, req, "/login", http.StatusSeeOther)
}