package main

import (

	"fmt"
	"html/template"
	"net/http"
	"os"
	"encoding/json"
	"strings"
	"strconv"

	"github.com/apexskier/httpauth"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"buscaminas2p"
)

var (

	backend     httpauth.LeveldbAuthBackend
	aaa         httpauth.Authorizer
	roles       map[string]httpauth.Role
	port        = 8009
	backendfile = "auth.leveldb.buscaminas"

	games     map[string]*buscaminas2p.Buscaminas
	players   map[int][2]string       
	idGame    int = 0

)

func main() {

	var err error
	os.Mkdir(backendfile, 0755)
	defer os.Remove(backendfile)

	// create the backend
	backend, err = httpauth.NewLeveldbAuthBackend(backendfile)
	if err != nil {
		panic(err)
	}

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

	r := mux.NewRouter()
	r.HandleFunc("/", handlePage)
	r.HandleFunc("/login", getLogin).Methods("GET")
	r.HandleFunc("/login", postLogin).Methods("POST")
	r.HandleFunc("/register", postRegister).Methods("POST")
	r.HandleFunc("/logout", handleLogout)

	r.HandleFunc("/lobby", handleLobby).Methods("POST")

	r.HandleFunc("/game/move", handleGameMove).Methods("POST")
	r.HandleFunc("/game/init", handleGameInit).Methods("POST")
	r.HandleFunc("/game/newGame", handleGameNewGame)
	r.HandleFunc("/game/joinGame", handleGameJoinGame)
	r.HandleFunc("/game/exit", handleGameExit)

	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("../images/"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("../js/"))))
	http.Handle("/fonts/", http.StripPrefix("/fonts/", http.FileServer(http.Dir("../fonts/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("../css/"))))
	http.Handle("/", r)

	fmt.Printf("Server running on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}


func handlePage(rw http.ResponseWriter, req *http.Request) {

	if err := aaa.Authorize(rw, req, true); err != nil {
		fmt.Println(err)
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
		return
	}

	if user, err := aaa.CurrentUser(rw, req); err == nil {

		t, err := template.ParseFiles("../views/lobby.html")
		if err != nil {
			panic (err)
		}

		t.Execute(rw, user)

	} else {
		panic(err)
	}

}

/******************************/
/*          Lobby             */
/******************************/

func handleLobby(rw http.ResponseWriter, req *http.Request) {
	

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

	fmt.Println(resp)

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

	http.Redirect(rw, req, "/", http.StatusSeeOther)	
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
		fmt.Println("Ya tiene dos jugadores.");
		return
	}

	players[gameId] = thisGame.Players

	games[user.Username] = thisGame
	http.Redirect(rw, req, "/game/newGame", http.StatusSeeOther)
}

func handleGameNewGame(rw http.ResponseWriter, req *http.Request) {

	user, err := aaa.CurrentUser(rw, req)
	if (err != nil) {
		panic(err)
	} else {

		t, err := template.ParseFiles("../views/game.html")	
		if err != nil {
			panic(err)
		}

		if _, ok := games[user.Username]; ok == true {
			fmt.Println("Ya está jugando.")
		} else {
			var new_game buscaminas2p.Buscaminas

			games[user.Username] = &new_game
			games[user.Username].Init(10, 15, 31, user.Username, idGame)
			players[idGame] = [2]string{user.Username, ""}
			idGame++
			
			fmt.Println("Juego creado.")
			games[user.Username].PrintBoard()
		}

		t.Execute(rw, user)
	}
	
}

func handleGameMove(rw http.ResponseWriter, req *http.Request) {

	//fmt.Println(req.FormValue("move"))

	w    := strings.Split(req.FormValue("move"), "_")
	x, _ := strconv.Atoi(w[0])
	y, _ := strconv.Atoi(w[1])
	fmt.Println(x, y)

	user, err := aaa.CurrentUser(rw, req)
	if err != nil {
		panic(err)
	}

	thisGame := games[user.Username]

	if thisGame.Players[thisGame.Turn] != user.Username {
		fmt.Fprint(rw, "null")
		return
	}

	resp := thisGame.Move(int16(x), int16(y))
	//games[user.Username].PrintStateBoard()

	respJson, _ := json.Marshal(resp)
	fmt.Fprint(rw, string(respJson))
}

func handleGameInit(rw http.ResponseWriter, req *http.Request) {

	user, err := aaa.CurrentUser(rw, req)
	if err != nil {
		panic(err)
	}

	thisGame := games[user.Username]

	board      := thisGame.Board
	stateBoard := thisGame.StateBoard
	r, c       := thisGame.R, thisGame.C
	turn     := thisGame.Turn
	score      := thisGame.Score
	minesLeft  := thisGame.MinesLeft
	players    := thisGame.Players

	type Response struct {

		Board [][]int16
		StateBoard [][]int16
		Turn int16
		R, C int16
		Score [2]int16
		Mines int16
		Username string
		Players [2]string
	}

	resp := Response{Board:board, StateBoard:stateBoard, R:r, C:c, 
					Turn:turn, Score:score, Mines:minesLeft, Username:user.Username, 
					Players:players}
	respJson, _ := json.Marshal(resp)

	fmt.Fprint(rw, string(respJson))
}

/******************************/
/*          Session           */
/******************************/

func getLogin(rw http.ResponseWriter, req *http.Request) {

	if _, err := aaa.CurrentUser(rw, req); err == nil {
		http.Redirect(rw, req, "/", http.StatusSeeOther)
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
	if err := aaa.Login(rw, req, username, password, "/"); err != nil && err.Error() == "already authenticated" {
		http.Redirect(rw, req, "/", http.StatusSeeOther)
	} else if err != nil {
		fmt.Println(err)
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
	}
}

func postRegister(rw http.ResponseWriter, req *http.Request) {
	var user httpauth.UserData
	user.Username = req.PostFormValue("username")
	user.Email = req.PostFormValue("email")
	password := req.PostFormValue("password")
	if err := aaa.Register(rw, req, user, password); err == nil {
		getLogin(rw, req)
	} else {
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
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

	http.Redirect(rw, req, "/", http.StatusSeeOther)
}