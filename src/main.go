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

	"github.com/flagz/src/buscaminas2p"
)

type myHandler struct {}

type myMux map[string]map[string]func(http.ResponseWriter, *http.Request)

func (m *myMux) addRoute(path string, f func(http.ResponseWriter, *http.Request), methods []string) {

	for i := range methods {
		_, ok := (*m)[methods[i]]
		if ok == false {
			(*m)[methods[i]] = make(map[string]func(http.ResponseWriter, *http.Request))
		}
		(*m)[methods[i]][path] = f
	}
}

var (

	backend     httpauth.SqlAuthBackend
	aaa         httpauth.Authorizer
	roles       map[string]httpauth.Role
	port        = 8009
	backenddb   = "david:david123@tcp(127.0.0.1:3306)/flagz?parseTime=true&loc=Local"

	games     map[string]*buscaminas2p.Buscaminas
	players   map[int][2]string       
	idGame    int = 0
	
	mux       myMux
	filehttp  = http.NewServeMux()
)

func main() {

	var err error

	// create the backend
	backend, err = httpauth.NewSqlAuthBackend("mysql", backenddb)
	if err != nil {
		panic(err)
	}
	defer backend.Close()

	games     = make(map[string]*buscaminas2p.Buscaminas)
	players   = make(map[int][2]string)

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

	mux.addRoute("/game", handleGame, []string{"GET", "POST"})
	mux.addRoute("/game/move", handleGameMove, []string{"GET", "POST"})
	mux.addRoute("/game/data", handleGameData, []string{"GET", "POST"})
	mux.addRoute("/game/joinGame", handleGameJoinGame, []string{"GET", "POST"})
	mux.addRoute("/game/exit", handleGameExit, []string{"GET", "POST"})

	filehttp.Handle("/", http.FileServer(http.Dir("../")))

	fmt.Printf("Server running on port %d\n", port)

	var mh *myHandler
	go CleanLoggedUsers()
	http.ListenAndServe(fmt.Sprintf(":%d", port), mh)
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

func (*myHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	//fmt.Println(req.URL.Path, req.Method)

	if (strings.Contains(req.URL.Path, ".")) {
		filehttp.ServeHTTP(rw, req)
	}

	if f, ok := mux[req.Method][req.URL.Path]; ok {

		user, err := aaa.CurrentUser(rw, req)
		if err != nil && req.URL.Path != "/login" && req.URL.Path != "/register" {

			http.Redirect(rw, req, "/login", http.StatusSeeOther)
		} else {

			if (err == nil) {
				lastTime, _  := backend.GetLastSeen(user.Username)

				duration := time.Since(lastTime)
			
				if duration.Minutes() > 5 {
					handleLogout(rw, req)
					return
				} else {
					backend.UpdateLogged(true, user.Username)
					backend.UpdateLastSeen(user.Username)
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
		if usernames[0] == "" || usernames[1] == "" {
			joinable = 1
		}
		resp[i] = myResp{GameId:gameId, Joinable:joinable, Players:usernames}
		i++
	}

	respJson, _ := json.Marshal(resp)
	fmt.Fprint(rw, string(respJson))
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
	fmt.Fprint(rw, string(respJson))
}


/******************************/
/*           Game             */
/******************************/

func deletePlayerFromGame(username string) {
	
	if _, ok := games[username]; ok == false {
		return
	}

	thisGame := games[username]
	gameId   := thisGame.Id

	delete(games, username)
	
	if players[gameId][0] == username {
		thisGame.Players[0] = ""
	} else if players[gameId][1] == username {
		thisGame.Players[1] = ""
	}
	players[gameId] = thisGame.Players

	if players[gameId][0] == "" && players[gameId][1] == "" {
		delete(players, gameId)
		thisGame = nil
	}
}

func handleGameExit(rw http.ResponseWriter, req *http.Request) {

	user, err := aaa.CurrentUser(rw, req)
	if (err != nil) {
		panic(err)
	}

	deletePlayerFromGame(user.Username)
}

func handleGameJoinGame(rw http.ResponseWriter, req *http.Request) {

	gameId, _ := strconv.Atoi(req.FormValue("id"))
	var versus string
	if players[gameId][0] != "" {
		versus = players[gameId][0]
	} else {
		versus = players[gameId][1]
	}
	thisGame  := games[versus]
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

	games[user.Username] = thisGame
	http.Redirect(rw, req, "/game", http.StatusSeeOther)
}

func handleGame(rw http.ResponseWriter, req *http.Request) {

	user, err := aaa.CurrentUser(rw, req)
	if (err != nil) {
		panic(err)
	} else {

		t, err := template.ParseFiles("../views/game.html")	
		if err != nil {
			panic(err)
		}

		if _, ok := games[user.Username]; ok == true {
			fmt.Println("It's playing already.")
		} else {
			var new_game buscaminas2p.Buscaminas

			games[user.Username] = &new_game
			games[user.Username].Init(10, 15, 31, user.Username, idGame)
			players[idGame] = [2]string{user.Username, ""}
			idGame++
			
			fmt.Println("Game created.")
			games[user.Username].PrintBoard()
		}

		t.Execute(rw, user)
	}
	
}

func handleGameMove(rw http.ResponseWriter, req *http.Request) {

	//fmt.Println(req.FormValue("move"))

	user, err := aaa.CurrentUser(rw, req)
	if err != nil {
		panic(err)
	}

	thisGame := games[user.Username]

	if thisGame.Players[thisGame.Turn] != user.Username {
		fmt.Fprint(rw, "null")
		return
	}

	usedBomb, _ := strconv.ParseBool(req.FormValue("usedBomb"))
	ids         := strings.Split(req.FormValue("move"), ",")

	var coord [][2]int16 = make([][2]int16, 0)
	for i := range ids {
		id   := strings.Split(ids[i], "_")
		x, _ := strconv.Atoi(id[0])
		y, _ := strconv.Atoi(id[1])
		coord = append(coord, [2]int16{int16(x), int16(y)})
	}

	resp := thisGame.Move(coord, usedBomb)
	//games[user.Username].PrintStateBoard()

	respJson, _ := json.Marshal(resp)
	fmt.Fprint(rw, string(respJson))
}

func handleGameData(rw http.ResponseWriter, req *http.Request) {

	user, err := aaa.CurrentUser(rw, req)
	if err != nil {
		panic(err)
	}

	if _, ok := games[user.Username]; ok == false {
		fmt.Println("The user has no game.")
		return
	}

	thisGame     := games[user.Username]

	board        := thisGame.Board
	stateBoard   := thisGame.StateBoard
	r, c         := thisGame.R, thisGame.C
	turn         := thisGame.Turn
	score        := thisGame.Score
	minesLeft    := thisGame.MinesLeft
	players      := thisGame.Players
	lastX, lastY := thisGame.LastX, thisGame.LastY
	hasBomb      := thisGame.HasBomb

	type Response struct {

		Board [][]int16
		StateBoard [][]int16
		Turn int16
		R, C int16
		LastX, LastY int16
		Score [2]int16
		Mines int16
		Username string
		Players [2]string
		HasBomb [2]bool
	}

	resp := Response{Board:board, StateBoard:stateBoard, R:r, C:c, 
					Turn:turn, Score:score, Mines:minesLeft, Username:user.Username, 
					Players:players, LastX:lastX, LastY:lastY, HasBomb:hasBomb}
	respJson, _ := json.Marshal(resp)

	fmt.Fprint(rw, string(respJson))
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