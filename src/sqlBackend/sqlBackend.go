package httpauth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	
)

// SqlAuthBackend database and database connection information.
type SqlAuthBackend struct {
	driverName     string
	dataSourceName string
	db             *sql.DB

	// prepared statements
	userStmt   *sql.Stmt
	usersStmt  *sql.Stmt
	insertStmt *sql.Stmt
	updateStmt *sql.Stmt
	deleteStmt *sql.Stmt

	playerStmt  *sql.Stmt
	playersStmt *sql.Stmt
	insertPlayerStmt *sql.Stmt
	updatePlayerStmt *sql.Stmt
	deletePlayerStmt *sql.Stmt
	updateLastSeenPlayerStmt *sql.Stmt
	lastSeenPlayerStmt *sql.Stmt
}

func mksqlerror(msg string) error {
	return errors.New("sqlbackend: " + msg)
}

// NewSqlAuthBackend initializes a new backend by testing the database
// connection and making sure the storage table exists. The table is called
// goauth.
//
// Returns an error if connecting to the database fails, pinging the database
// fails, or creating the table fails.
//
// This uses the databases/sql package to open a connection. Its parameters
// should match the sql.Open function. See
// http://golang.org/pkg/database/sql/#Open for more information.
//
// Be sure to import "database/sql" and your driver of choice. If you're not
// using sql for your own purposes, you'll need to use the underscore to import
// for side effects; see http://golang.org/doc/effective_go.html#blank_import.
func NewSqlAuthBackend(driverName, dataSourceName string) (b SqlAuthBackend, e error) {
	b.driverName = driverName
	b.dataSourceName = dataSourceName
	if driverName == "sqlite3" {
		if _, err := os.Stat(dataSourceName); os.IsNotExist(err) {
			return b, ErrMissingBackend
		}
	}
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return b, mksqlerror(err.Error())
	}
	err = db.Ping()
	if err != nil {
		return b, mksqlerror(err.Error())
	}
	b.db = db
	_, err = db.Exec(`create table if not exists goauth (Username varchar(255), 
														Email varchar(255), 
														Hash varchar(255), 
														Role varchar(255), 
														primary key (Username))`)
	if err != nil {
		return b, mksqlerror(err.Error())
	}
	_, err = db.Exec(`create table if not exists player (Username varchar(255), 
														Logged boolean, 
														GamesPlayed int, 
														GamesWon int, 
														LastSeen datetime default NOW(), 
														primary key (Username), 
														foreign key (Username) references goauth(Username))`)
	if err != nil {
		return b, mksqlerror(err.Error())
	}

	// prepare statements for concurrent use and better preformance
	//
	// NOTE:
	// I don't want to have to check if it's postgres, but postgres uses
	// different tokens for placeholders. :( Also be aware that postgres
	// lowercases all these column names.
	//
	// Thanks to mjhall for letting me know about this.
	if driverName == "postgres" {
		b.userStmt, err = db.Prepare(`select Email, Hash, Role from goauth where Username = $1`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("userstmt: %v", err))
		}
		b.usersStmt, err = db.Prepare(`select Username, Email, Hash, Role from goauth`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("usersstmt: %v", err))
		}
		b.insertStmt, err = db.Prepare(`insert into goauth (Username, Email, Hash, Role) values ($1, $2, $3, $4)`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("insertstmt: %v", err))
		}
		b.updateStmt, err = db.Prepare(`update goauth set Email = $1, Hash = $2, Role = $3 where Username = $4`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("updatestmt: %v", err))
		}
		b.deleteStmt, err = db.Prepare(`delete from goauth where Username = $1`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("deletestmt: %v", err))
		}
	} else {
		b.userStmt, err = db.Prepare(`select Email, Hash, Role from goauth where Username = ?`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("userstmt: %v", err))
		}
		b.usersStmt, err = db.Prepare(`select Username, Email, Hash, Role from goauth`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("usersstmt: %v", err))
		}
		b.insertStmt, err = db.Prepare(`insert into goauth (Username, Email, Hash, Role) values (?, ?, ?, ?)`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("insertstmt: %v", err))
		}
		b.updateStmt, err = db.Prepare(`update goauth set Email = ?, Hash = ?, Role = ? where Username = ?`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("updatestmt: %v", err))
		}
		b.deleteStmt, err = db.Prepare(`delete from goauth where Username = ?`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("deletestmt: %v", err))
		}

		/**** Player Statements ****/
		b.playerStmt, err = db.Prepare(`select Username, Logged, GamesPlayed, GamesWon, LastSeen from player where Username = ?`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("playerstmt: %v", err))
		}
		b.playersStmt, err = db.Prepare(`select Username, Logged, GamesPlayed, GamesWon, LastSeen from player`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("playersstmt: %v", err))
		}
		b.insertPlayerStmt, err = db.Prepare(`insert into player (Username, Logged, GamesPlayed, GamesWon) values (?, ?, ?, ?)`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("insertPlayerstmt: %v", err))
		}
		b.updatePlayerStmt, err = db.Prepare(`update player set Username = ?, Logged = ?, GamesPlayed = ?, GamesWon = ?`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("updatePlayerstmt: %v", err))
		}
		b.deletePlayerStmt, err = db.Prepare(`delete from player where Username = ?`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("deletePlayerstmt: %v", err))
		}
		b.updateLastSeenPlayerStmt, err = db.Prepare(`update player set LastSeen = NOW() where Username = ?`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("updateLastSeenPlayerstmt: %v", err))
		}
		b.lastSeenPlayerStmt, err = db.Prepare(`select LastSeen from player where Username = ?`)
		if err != nil {
			return b, mksqlerror(fmt.Sprintf("lastSeenPlayerstmt: %v", err))
		}
	}

	return b, nil
}

// User returns the user with the given username. Error is set to
// ErrMissingUser if user is not found.
func (b SqlAuthBackend) User(username string) (user UserData, e error) {
	row := b.userStmt.QueryRow(username)
	err := row.Scan(&user.Email, &user.Hash, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrMissingUser
		}
		return user, mksqlerror(err.Error())
	}

	user.Username = username
	return user, nil
}

// Users returns a slice of all users.
func (b SqlAuthBackend) Users() (us []UserData, e error) {
	rows, err := b.usersStmt.Query()
	if err != nil {
		return us, mksqlerror(err.Error())
	}
	var (
		username, email, role string
		hash                  []byte
	)
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&username, &email, &hash, &role)
		if err != nil {
			return us, mksqlerror(err.Error())
		}
		us = append(us, UserData{username, email, hash, role})
	}
	return us, nil
}

// SaveUser adds a new user, replacing one with the same username.
func (b SqlAuthBackend) SaveUser(user UserData) (err error) {
	if _, err := b.User(user.Username); err == nil {
		_, err = b.updateStmt.Exec(user.Email, user.Hash, user.Role, user.Username)
	} else {
		_, err = b.insertStmt.Exec(user.Username, user.Email, user.Hash, user.Role)
		err = b.SavePlayer(PlayerData{user.Username, false, 0, 0})
	}
	return
}

// DeleteUser removes a user, raising ErrDeleteNull if that user was missing.
func (b SqlAuthBackend) DeleteUser(username string) error {
	result, err := b.deleteStmt.Exec(username)
	if err != nil {
		return mksqlerror(err.Error())
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return mksqlerror(err.Error())
	}
	if rows == 0 {
		return ErrDeleteNull
	}
	return nil
}

/***********************/
/*       Player        */
/***********************/

func (b SqlAuthBackend) Player(username string) (PlayerData, error) {

	var player PlayerData

	row := b.playerStmt.QueryRow(username)
	err := row.Scan(&player.Logged, &player.GamesPlayed, &player.GamesWon)
	if err != nil {
		if err == sql.ErrNoRows {
			return player, ErrMissingUser
		}
		return player, mksqlerror(err.Error())
	}

	player.Username = username
	return player, nil
}

func (b SqlAuthBackend) Players() ([]PlayerData, error) {

	var players []PlayerData

	rows, err := b.playersStmt.Query()
	if err != nil {
		return players, mksqlerror(err.Error())
	}
	var (
		username string
		logged   bool
		gamesPlayed, gamesWon int
	)
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&username, &logged, &gamesPlayed, &gamesWon)
		if err != nil {
			return players, mksqlerror(err.Error())
		}
		players = append(players, PlayerData{username, logged, gamesPlayed, gamesWon})
	}
	return players, nil
}

func (b SqlAuthBackend) SavePlayer(player PlayerData) error {
	_, err := b.Player(player.Username)
	if err == nil {
		_, err = b.updatePlayerStmt.Exec(player.Logged, player.GamesPlayed, player.GamesWon, player.Username)
	} else {
		_, err = b.insertPlayerStmt.Exec(player.Username, player.Logged, player.GamesPlayed, player.GamesWon)
	}
	return err
}

func (b SqlAuthBackend) DeletePlayer(username string) error {
	result, err := b.deletePlayerStmt.Exec(username)
	if err != nil {
		return mksqlerror(err.Error())
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return mksqlerror(err.Error())
	}
	if rows == 0 {
		return ErrDeleteNull
	}
	return nil
}

func (b SqlAuthBackend) UpdateLastSeen(username string) error {
	_, err := b.updateLastSeenPlayerStmt.Exec(username)
	return err;
}

func (b SqlAuthBackend) GetLastSeen(username string) (time.Time, error) {
	var time time.Time

	row := b.lastSeenPlayerStmt.QueryRow(username)
	err := row.Scan(&time)
	if err != nil {
		if err == sql.ErrNoRows {
			return time, ErrMissingUser
		}
		return time, mksqlerror(err.Error())
	}

	return time, nil
}

// Close cleans up the backend by terminating the database connection.
func (b SqlAuthBackend) Close() {
	b.db.Close()
	b.userStmt.Close()
	b.usersStmt.Close()
	b.insertStmt.Close()
	b.updateStmt.Close()
	b.deleteStmt.Close()

	b.playerStmt.Close()
	b.playersStmt.Close()
	b.insertPlayerStmt.Close()
	b.updatePlayerStmt.Close()
	b.deletePlayerStmt.Close()
}
