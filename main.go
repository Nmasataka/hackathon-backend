package main

import (
	"encoding/json"
	"github.com/oklog/ulid"
	"hackathon-backend/database"
	"hackathon-backend/handler"
	"hackathon-backend/models"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/oklog/ulid/v2"
)

var entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)

func posttweet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	//w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodGet:
		// ②-1
		name := r.URL.Query().Get("name") // To be filled
		if name == "" {
			log.Println("fail: name is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// ②-2
		rows, err := database.Db.Query("SELECT id, name, age FROM user WHERE name = ?", name)
		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// ②-3
		users := make([]models.UserResForHTTPGet, 0)
		for rows.Next() {
			var u models.UserResForHTTPGet
			if err := rows.Scan(&u.Id, &u.Name, &u.Age); err != nil {
				log.Printf("fail: rows.Scan, %v\n", err)

				if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
					log.Printf("fail: rows.Close(), %v\n", err)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			users = append(users, u)
		}

		// ②-4
		bytes, err := json.Marshal(users)
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)
	case http.MethodPost:
		var req models.TweetForHTTPPOST
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Printf("fail: json.NewDecoder.Decode, %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("ポストはされている")

		tx, err := database.Db.Begin()
		if err != nil {
			log.Printf("fail: db.Begin, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec("INSERT INTO Tweet ( uid, content) VALUES ( ?, ?)", req.Uid, req.Content)
		if err != nil {
			log.Printf("fail: tx.Exec, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := tx.Commit(); err != nil {
			log.Printf("fail: tx.Commit, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		log.Printf("kakuninn%s", req.Uid)

		resp := map[string]string{"id": req.Uid}
		bytes, err := json.Marshal(resp)
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	default:
		log.Printf("fail: HTTP Method is %s\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func getAllTweets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodGet:
		uid := r.URL.Query().Get("uid") // To be filled
		postuid := r.URL.Query().Get("postuid")
		if uid == "" || postuid == "" {
			log.Println("fail: uid is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var condition string
		var args []interface{}

		if postuid == "all" {
			condition = "1 = 1"
		} else {
			condition = "t.uid = ?"
			args = append(args, postuid)
		}

		query := `SELECT 
            t.tweet_id,
            t.uid AS profile_uid,
			u.username AS username,
            t.content,
            t.created_at,
            t.likes_count,
			t.retweet_count,
            CASE WHEN l.uid IS NOT NULL THEN TRUE ELSE FALSE END AS liked_by_user
        FROM 
            Tweet t
        LEFT JOIN 
            Likes l
        ON 
            t.tweet_id = l.tweet_id AND l.uid = ?
		LEFT JOIN 
			User u
		ON 
			t.uid = u.uid
        WHERE 
            ` + condition + `
        ORDER BY 
            t.created_at DESC;
    `

		args = append([]interface{}{uid}, args...)
		//rows, err := database.Db.Query("SELECT tweet_id,uid,content,created_at,likes_count FROM Tweet WHERE uid = ?", uid)
		rows, err := database.Db.Query(query, args...)
		/*
					rows, err := database.Db.Query(`
			        SELECT
			            t.tweet_id,
			            t.uid AS profile_uid,
			            t.content,
			            t.created_at,
			            t.likes_count,
						t.retweet_count,
			            CASE WHEN l.uid IS NOT NULL THEN TRUE ELSE FALSE END AS liked_by_user
			        FROM
			            Tweet t
			        LEFT JOIN
			            Likes l
			        ON
			            t.tweet_id = l.tweet_id AND l.uid = ?
			        WHERE
			            t.uid = ?
			        ORDER BY
			            t.created_at DESC;
			    `, uid, postuid)

		*/
		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tweets := make([]models.TweetWithLikeStatus, 0)
		for rows.Next() {
			var u models.TweetWithLikeStatus
			var createdAt []byte // まずバイト列で受け取る
			if err := rows.Scan(&u.Tweet_id, &u.Uid, &u.Username, &u.Content, &createdAt, &u.Likes_count, &u.Retweet_count, &u.IsLiked); err != nil {
				log.Printf("fail: rows.Scan, %v\n", err)

				if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
					log.Printf("fail: rows.Close(), %v\n", err)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// バイト列から time.Time に変換
			u.Created_at, err = time.Parse("2006-01-02 15:04:05", string(createdAt)) // 必要に応じてフォーマットを変更
			if err != nil {
				log.Printf("fail: time.Parse, %v\n", err)
				rows.Close()
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			tweets = append(tweets, u)
		}

		bytes, err := json.Marshal(tweets)
		if err != nil {
			log.Printf("fail: json.Marshal, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
	default:
		log.Printf("fail: HTTP Method is %s\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func main() {
	if err := database.InitDB(); err != nil {
		log.Fatal("Error initializing database: ", err)
	}

	// ② /userでリクエストされたらnameパラメーターと一致する名前を持つレコードをJSON形式で返す
	http.HandleFunc("/user", handler.Handler)
	http.HandleFunc("/users", handler.GetAllUsers)
	http.HandleFunc("/register", handler.Register)
	http.HandleFunc("/posttweet", posttweet)
	http.HandleFunc("/tweetlist", getAllTweets)

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
