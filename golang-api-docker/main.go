package main

import (
    "database/sql"
    "log"
    "os"
//    "fmt"
    "net/http"
    "strings"
    _ "github.com/go-sql-driver/mysql"
)

type Post struct {
    ID   string    `json:"id"`
    Title string `json:"title"`
    Author string `json:"author"`
    Date string `json:"date"`
}

func postHandler(post Post) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(post.ID+"\n"))
		w.Write([]byte(post.Title+"\n"))
		w.Write([]byte(post.Author+"\n"))
		w.Write([]byte(post.Date+"\n"))
	}
	return http.HandlerFunc(fn)
}

func postHandlerAll(posts string) http.Handler {
        fn := func(w http.ResponseWriter, r *http.Request) {
                w.Write([]byte(posts+"\n"))
        }
        return http.HandlerFunc(fn)
}

func main() {
    var ph http.Handler
    var phAll http.Handler
    var psAll strings.Builder

    // Open up our database connection.
    
     db_host := os.Getenv("DB_HOST")
     db_password := os.Getenv("DB_PASSWORD")
     db_user := os.Getenv("DB_USER")
     db_name := os.Getenv("DB_NAME")
     db_port := os.Getenv("DB_PORT")

    db, err := sql.Open("mysql",db_user+":"+db_password+"@tcp("+db_host+":"+db_port+")/"+db_name)

    // if there is an error opening the connection, handle it
    if err != nil {
        log.Print(err.Error())
    }
    defer db.Close()

    // Execute the query
    results, err := db.Query("select id, post_title, post_author, post_date from wp_posts")
    if err != nil {
        panic(err.Error()) // proper error handling instead of panic in your app
    }

    for results.Next() {
        var post Post
        // for each row, scan the result into our tag composite object
        err = results.Scan(&post.ID, &post.Title,  &post.Author, &post.Date)
        if err != nil {
            panic(err.Error()) // proper error handling instead of panic in your app
        }
                // and then print out the tag's Name attribute
  //      fmt.Println(post.ID+" "+post.Title+" "+post.Author+" "+post.Date)
	psAll.WriteString(post.ID+" "+post.Title+" "+post.Author+" "+post.Date+"\n")
	ph = postHandler(post)
        http.Handle("/posts/post/"+post.ID, ph)
    }
    
    phAll = postHandlerAll(psAll.String()) 
    http.Handle("/posts", phAll)
    //http.Handle("/posts/post/"+post.ID, ph)
    http.ListenAndServe(":3000", nil)   
}
