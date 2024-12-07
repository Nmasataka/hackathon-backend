package handler

import (
	"encoding/json"
	"hackathon-backend/database"
	"hackathon-backend/models"
	"log"
	"net/http"
	"time"
)

func Replytweet(w http.ResponseWriter, r *http.Request) {
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
		var req models.ReplytweetForHTTPPOST
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

		_, err = tx.Exec("INSERT INTO Reply ( uid, content, parent_tweet_id) VALUES ( ?, ?, ?)", req.Uid, req.Content, req.Parent_tweet_id)
		if err != nil {
			log.Printf("fail: tx.Exec, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = tx.Exec(`
            UPDATE Tweet SET retweet_count = retweet_count + 1  WHERE tweet_id = ?`, req.Parent_tweet_id)
		if err != nil {
			log.Printf("fail: update likes_count, %v\n", err)
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

func GetAllReplyTweets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodGet:
		uid := r.URL.Query().Get("uid") // To be filled
		id := r.URL.Query().Get("parent_tweet_id")
		if uid == "" || id == "" {
			log.Println("fail: uid is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//rows, err := database.Db.Query(`select reply_id, uid, content,created_at,likes_count
		//from Reply where parent_tweet_id = ?`, id)

		rows, err := database.Db.Query(`SELECT
			            t.reply_id,
			            t.uid AS profile_uid,
						u.username AS username,u.profile_picture,
			            t.content,
			            t.created_at,
			            t.likes_count,
			            CASE WHEN l.uid IS NOT NULL THEN TRUE ELSE FALSE END AS liked_by_user
			        FROM
			            Reply t
			        LEFT JOIN
			            ReplyLikes l
			        ON
			            t.reply_id = l.reply_id AND l.uid = ?
					LEFT JOIN 
						User u
					ON 
						t.uid = u.uid
			        WHERE
			            t.parent_tweet_id = ?
			        ORDER BY
			            t.created_at DESC;`, uid, id)

		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tweets := make([]models.ReplyListForHTTPGET, 0)
		for rows.Next() {
			var u models.ReplyListForHTTPGET
			var createdAt []byte // まずバイト列で受け取る
			if err := rows.Scan(&u.Reply_id, &u.Uid, &u.Username, &u.ProfilePicture, &u.Content, &createdAt, &u.Likes_count, &u.IsLiked); err != nil {
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
