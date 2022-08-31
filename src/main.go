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
	//url := "http://abehiroshi.la.coocan.jp/"

	//url := "https://www.hirorocafe.com/"

	url := "https://lumine.yoshimoto.co.jp/schedule/"

	maxConnection := make(chan bool, 200)
	wg := &sync.WaitGroup{}

	count := 0
	start := time.Now()

	for maxRequest := 0; maxRequest < 1; maxRequest++ {
		wg.Add(1)
		maxConnection <- true
		go func() { // go func(){/*処理*/}とやると並列処理を開始してくれるよ！
			defer wg.Done() // wg.Done()を呼ぶと並列処理が一つ終わったことを便利な奴に教えるよ！

			// resp, err := http.Get(url) // GETリクエストでアクセスするよ！
			// if err != nil {
			// 	fmt.Fprintf(os.Stderr, "print: %v\n", err)
			// 	return // 回線が狭かったりするとここでエラーが帰ってくるよ！
			// }
			// defer resp.Body.Close() // 関数が終了するとなんかクローズするよ！（おまじない的な）

			// body, _ := ioutil.ReadAll(resp.Body)

			// // 文字コードを判定
			// detector := chardet.NewTextDetector()
			// detectResult, _ := detector.DetectBest(content)
			// // => 例：UTF-8

			// 文字コードの変換
			// bufferReader := bytes.NewReader(content)
			// reader, _ := charset.NewReaderLabel(detectResult.Charset, bufferReader)
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

			// HTMLパース
			doc, _ := goquery.NewDocumentFromReader(reader)
			//fmt.Fprintf(w, "<h2>パースされた文字コード：%#v</h2>", doc)

			//rslt := doc.Find("div.container > p")
			rslt := doc.Find("div.schedule-time")
			rslt.Each(func(i int, s *goquery.Selection) {
				regex := regexp.MustCompile(`.*オズワルド.*`)
				res := regex.MatchString(s.Text())
				// if res == true {
				// 	parentSelection := s.Parent()
				// 	ps, _ := parentSelection.Html()
				// 	fmt.Fprintf(w, "親セレクション：%s", ps)
				// 	scheduleDetailSelection := parentSelection.ChildrenFiltered("schedule-detail")
				// 	selections, err := scheduleDetailSelection.Html()
				// 	if err != nil {
				// 		fmt.Fprintln(w, "セレクションなんてありません", err)
				// 	} else {
				// 		fmt.Fprintf(w, "セレクションの一覧：%s\n", selections)
				// 		fmt.Fprintf(w, "セレクション：%s\n", scheduleDetailSelection.Text())
				// 	}
				// btns := scheduleDetailSelection.ChildrenFiltered("btns > ul > li")
				// if btns == nil {
				// 	fmt.Fprintln(w, "btnsセクションがないです。")
				// }
				// val, exists := btns.Attr("href")
				// if !exists {
				// 	fmt.Fprintln(w, "リンクがないです。", val)
				// } else {
				// 	fmt.Fprintln(w, "scheduleの値：%s</h2>", val)
				// }
				// } else {
				// 	fmt.Fprintln(w, "かまいたちは存在しません。")
				// }
				fmt.Fprintf(w, "<h2>%f</h2>", res)
			})

			//fmt.Fprintf(w, "<h2>文字コード：%f</h2>", detectResult.Charset)

			count++         // アクセスが成功したことをカウントするよ！
			<-maxConnection // ここは並列する数を抑制する奴だよ！詳しくはググって！
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
