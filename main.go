package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var DbConnection *sql.DB

type Page struct { //情報を受け取る構造体を定義
	Title string
	Body  []byte
}

type Person struct {
	Name string
	Age  int
}

func (p *Page) save() error { //情報をファイルに保存するメソッドを定義
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	//Docker Container内でtxtファイルを検索するときは以下のコードでbuildする。
	//body, err := ioutil.ReadFile("/var/www/" + filename)
	if err != nil {
		log.Fatal(err)
	}
	return &Page{Title: title, Body: body}, nil

}

func viewHundler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/view/"):]
	p, err := loadPage(title)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(w, "<h1>%s</h1><div>%s<div>", p.Title, p.Body)

}

func main() {
	// DbConnection, _ := sql.Open("sqlite3", "./example.sql")
	// defer DbConnection.Close()

	// cmd := `CREATE TABLE IF NOT EXISTS person(
	// 	name STRING,
	// 	age INT)`
	// _, err := DbConnection.Exec(cmd)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// cmd = "INSERT INTO person (name, age) VALUES (?, ?)"
	// _, err = DbConnection.Exec(cmd, "Jirorian", 19) //SQLコマンド（cmd）のVALUESの？に値を入れている。
	// //?を使うのはsqlインジェクション対策
	// if err != nil {
	// 	log.Fatal(err)
	// }

	http.HandleFunc("/view/", viewHundler)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
