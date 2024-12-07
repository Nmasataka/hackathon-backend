package handler

import (
	"encoding/json"
	"hackathon-backend/database"
	"hackathon-backend/models"
	"log"
	"net/http"
	"time"
)

func GetTweet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodGet:
		id := r.URL.Query().Get("tweet_id") // To be filled
		uid := r.URL.Query().Get("uid")

		if id == "" || uid == "" {
			log.Println("fail: uid is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//rows, err := database.Db.Query("SELECT tweet_id,uid,content,created_at,likes_count FROM Tweet WHERE tweet_id = ?", id)
		rows, err := database.Db.Query(`
    SELECT 
        t.tweet_id,
        t.uid AS profile_uid,
		u.username AS username,
		u.profile_picture,
        t.content,
        t.created_at,
        t.likes_count,
		t.retweet_count,t.image_url,
        CASE WHEN l.uid IS NOT NULL THEN TRUE ELSE FALSE END AS liked_by_user
    FROM 
        Tweet t
    LEFT JOIN 
        Likes l
    ON 
        t.tweet_id = l.tweet_id AND l.uid = ?
	LEFT JOIN User u ON t.uid = u.uid
    WHERE 
        t.tweet_id = ?;
`, uid, id)

		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tweets := make([]models.TweetWithLikeStatus, 0)
		for rows.Next() {
			var u models.TweetWithLikeStatus
			var createdAt []byte // まずバイト列で受け取る
			if err := rows.Scan(&u.Tweet_id, &u.Uid, &u.Username, &u.ProfilePicture, &u.Content, &createdAt, &u.Likes_count, &u.Retweet_count, &u.Image_url, &u.IsLiked); err != nil {
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

func GetAllTweets(w http.ResponseWriter, r *http.Request) {
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
			u.profile_picture,
            t.content,
            t.created_at,
            t.likes_count,
			t.retweet_count,t.image_url,
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

		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tweets := make([]models.TweetWithLikeStatus, 0)
		for rows.Next() {
			var u models.TweetWithLikeStatus
			var createdAt []byte // まずバイト列で受け取る
			if err := rows.Scan(&u.Tweet_id, &u.Uid, &u.Username, &u.ProfilePicture, &u.Content, &createdAt, &u.Likes_count, &u.Retweet_count, &u.Image_url, &u.IsLiked); err != nil {
				log.Printf("ここ")
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
