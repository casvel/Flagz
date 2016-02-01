# Flagz

My first web project.

A two players minesweeper game. Version 1.0.

Clone inside `github` in your `GOPATH`. Run `go run main.go` inside `github.com/flagz/src`.

You will need the next packages:

 - `go get github.com/gorilla/mux`
 
 - `go get github.com/go-sql-driver/mysql`
 - `go get github.com/apexskier/httpauth`. I changed some things here: 
   - Comment line 137-140 in `github.com/apexskier/httpauth/auth.go`. 
   - You must replace the file `github.com/apexskier/httpauth/sqlBackend.go` with `github.com/flagz/src/sqlBackend/sqlBackend.go`

 Must have installed `mysql`. Create a database named `flagz` and give all permissions to the user: `david:david123`. Or you could change the variable `backenddb` inside `main.go` 
