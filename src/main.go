package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sclevine/agouti"
)

//go get -u github.com/PuerkitoBio/goquery スクレイピングライブラリ
//go get github.com/sclevine/agouti スクレイピングのためのドライバー
//go get -u github.com/saintfish/chardet # 文字コードの判定用
//go get -u golang.org/x/net/html/charset # 文字コードの変換用

// goquery参考記事
// https://qiita.com/Yaruki00/items/b50e346551690b158a79
// https://pkg.go.dev/github.com/PuerkitoBio/goquery#section-readme
// https://undersourcecode.hatenablog.com/entry/2018/12/23/103324
// https://qiita.com/Azunyan1111/items/a1b6c58dc868814efb51
// https://qiita.com/mmm888/items/42383f967e44e633f0eb
//バッチとかslack連携含めた参考記事はこれ↓
// https://ceblog.mediba.jp/post/657126495256510464/go-%E3%81%A7%E3%82%B9%E3%82%AF%E3%83%AC%E3%82%A4%E3%83%94%E3%83%B3%E3%82%B0%E3%81%97%E3%81%9F-mediba-%E3%81%AE%E8%A8%98%E4%BA%8B%E3%82%92-slack-%E3%81%AB%E6%8A%95%E7%A8%BF%E3%81%99%E3%82%8B%E3%83%90%E3%83%83%E3%83%81%E3%82%92%E4%BD%9C%E3%81%A3%E3%81%9F%E8%A9%B1

//前提：
//HTMLのDOM構造
//<要素名 属性="属性値">
//属性→attributeのこと。hrefとかdivとかのこと。
//要素→classとかidとか
//属性値→クラス名とか
//p{				←このpがselector、
//	color: red;		←このcolorがプロパティ。redが値。
//}

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

func findLink(n *html.Node, w http.ResponseWriter) {

	if n.Type == html.ElementNode && n.Data == "div" {

		for _, a := range n.Attr {
			if a.Val == "schedule" {
				//if a.Val != "" {
				fmt.Fprintf(w, "<h2>リンク：%f</h2>", a.Val)
				fmt.Fprintf(w, "<h2>リンク：%f</h2>", n.FirstChild)
				break
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		findLink(c, w)
	}

}

func scraping(w http.ResponseWriter, r *http.Request) {
	//阿部寛のページでスクレイピング試したい場合↓
	//url := "http://abehiroshi.la.coocan.jp/"

	//なんか知らん人の動的なブログでスクレイピング試したい場合↓
	//url := "https://www.hirorocafe.com/"

	//ルミネザ吉本でスクレイピングする場合↓
	url := "https://lumine.yoshimoto.co.jp/schedule/"

	maxConnection := make(chan bool, 200)
	wg := &sync.WaitGroup{}

	count := 0
	start := time.Now()

	//以下、1スレッドの意味のないゴルーチン
	for maxRequest := 0; maxRequest < 1; maxRequest++ {
		wg.Add(1)
		maxConnection <- true
		go func() { // go func(){/*処理*/}とやると並列処理を開始してくれる。
			defer wg.Done() // wg.Done()を呼ぶと並列処理が一つ終わったことを便利な奴に教える。

			//Chromeのドライバーの設定
			driver := agouti.ChromeDriver()
			defer driver.Stop()

			err := driver.Start()
			if err != nil {
				log.Printf("Failed to start driver: %v", err)
			}

			page, err := driver.NewPage(agouti.Browser("chrome"))
			if err != nil {
				log.Printf("Failed to open page: %v", err)
			}

			err = page.Navigate(url)
			if err != nil {
				log.Printf("Failed to navigate: %v", err)
			}

			// contentにHTMLが入る
			content, err := page.HTML()
			if err != nil {
				log.Printf("Failed to get html: %v", err)
			}

			reader := strings.NewReader(content)
			doc, _ := goquery.NewDocumentFromReader(reader)

			//任意の芸人に当てはまる開演日のa要素を探しにいくコード
			rslt := doc.Find("div.schedule-time")
			rslt.Each(func(i int, s *goquery.Selection) {
				regex := regexp.MustCompile(`.*オズワルド.*`)
				res := regex.MatchString(s.Text())
				if res == true {
					parentSelection := s.Parent()
					link := parentSelection.Find("div.btns")
					fmt.Fprintln(w, "リンク：%f", link.Text())
					a := link.Find("a")
					val, _ := a.Attr("href")
					fmt.Fprintln(w, "リンクやで", val)
				}
				fmt.Fprintf(w, "<h2>%#v</h2>", res)
			})

			count++         // アクセスが成功したことをカウントする
			<-maxConnection // ここは並列する数を抑制する奴。詳しくはググる
		}()
	}
	wg.Wait()
	end := time.Now()

	fmt.Fprintf(w, "<h2>%d 回のリクエストに成功しました！</h2>\n", count)
	fmt.Fprintf(w, "<h2>%f 秒処理に時間がかかりました！\n</h2>", (end.Sub(start)).Seconds())
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
	http.HandleFunc("/scraping/", scraping)
	log.Fatal(http.ListenAndServe(":8080", nil))

}
