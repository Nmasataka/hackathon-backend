package main

import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/oklog/ulid/v2"
	"hackathon-backend/database"
	"hackathon-backend/handler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := database.InitDB(); err != nil {
		log.Fatal("Error initializing database: ", err)
	}

	// ② /userでリクエストされたらnameパラメーターと一致する名前を持つレコードをJSON形式で返す
	http.HandleFunc("/user", handler.Handler)
	http.HandleFunc("/users", handler.GetAllUsers)
	http.HandleFunc("/register", handler.Register)
	http.HandleFunc("/posttweet", handler.Posttweet)
	http.HandleFunc("/tweetlist", handler.GetAllTweets)

	http.HandleFunc("/like", handler.Like)
	http.HandleFunc("/replylike", handler.ReplyLike)
	http.HandleFunc("/tweet", handler.GetTweet)
	http.HandleFunc("/reply", handler.Replytweet)
	http.HandleFunc("/replylist", handler.GetAllReplyTweets)
	http.HandleFunc("/register-userinfo", handler.RegisterUserInfo)
	http.HandleFunc("/loginusername", handler.FetchUsername)

	http.HandleFunc("/follow", handler.Follow)
	http.HandleFunc("/followtweetlist", handler.GetAllFollowTweets)

	http.HandleFunc("/following", handler.GetAllFollowing)
	http.HandleFunc("/followers", handler.GetAllFollower)
	http.HandleFunc("/generate", handler.HandleGenerate)
	http.HandleFunc("/searchtweetlist", handler.GetAllSearchTweet)
	http.HandleFunc("/searchuserlist", handler.SearchUser)

	// ③ Ctrl+CでHTTPサーバー停止時にDBをクローズする
	closeDBWithSysCall()

	// 8000番ポートでリクエストを待ち受ける
	log.Println("Listening...")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	}
}

// ③ Ctrl+CでHTTPサーバー停止時にDBをクローズする
func closeDBWithSysCall() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Printf("received syscall, %v", s)

		if err := database.Db.Close(); err != nil {
			log.Fatal(err)
		}
		log.Printf("success: db.Close()")
		os.Exit(0)
	}()
}
