package handler

import (
	"encoding/json"
	"hackathon-backend/database"
	"hackathon-backend/models"
	"log"
	"net/http"
	"time"
)

func Follow(w http.ResponseWriter, r *http.Request) {
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
		rows, err := database.Db.Query("SELECT uid FROM userlist WHERE email = ?", name)
		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// ②-3
		users := make([]models.LoginForHTTPGET, 0)
		for rows.Next() {
			var u models.LoginForHTTPGET
			if err := rows.Scan(&u.Email); err != nil {
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
		var req models.FollowForHTTPPOST
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Printf("fail: json.NewDecoder.Decode, %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		tx, err := database.Db.Begin()
		if err != nil {
			log.Printf("fail: db.Begin, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var exist bool
		err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM Follow WHERE follower_uid = ? AND followed_uid = ?)`, req.Follower_uid, req.Followed_uid).Scan(&exist)

		if err != nil {
			log.Printf("fail: query for like check, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if exist {
			// 既存の「いいね」を削除
			_, err := tx.Exec(`
            DELETE FROM Follow WHERE follower_uid = ? AND followed_uid = ?`, req.Follower_uid, req.Followed_uid)
			if err != nil {
				log.Printf("fail: delete like, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// 「いいね数」を減少

			_, err = tx.Exec(`
				            UPDATE User SET follower_count = follower_count - 1 WHERE uid = ?`, req.Followed_uid)
			if err != nil {
				log.Printf("fail: update likes_count, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, err = tx.Exec(`
				            UPDATE User SET follow_count = follow_count - 1 WHERE uid = ?`, req.Follower_uid)
			if err != nil {
				log.Printf("fail: update follow_count, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		} else {
			// 新しい「いいね」を追加
			_, err := tx.Exec(`
            INSERT INTO Follow (follower_uid, followed_uid, followed_at) VALUES (?, ?, NOW())`, req.Follower_uid, req.Followed_uid)
			if err != nil {
				log.Printf("fail: insert like, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			_, err = tx.Exec(`
				            UPDATE User SET follower_count = follower_count + 1 WHERE uid = ?`, req.Followed_uid)
			if err != nil {
				log.Printf("fail: update likes_count, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			_, err = tx.Exec(`
				            UPDATE User SET follow_count = follow_count + 1 WHERE uid = ?`, req.Follower_uid)
			if err != nil {
				log.Printf("fail: update follow_count, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		}

		if err := tx.Commit(); err != nil {
			log.Printf("fail: tx.Commit, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		resp := map[string]string{"status": "success"}
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

func GetAllFollowTweets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodGet:
		uid := r.URL.Query().Get("uid") // To be filled
		if uid == "" {
			log.Println("fail: uid is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("%s", uid)

		rows, err := database.Db.Query(`
			        SELECT
			            t.tweet_id,
			            t.uid AS profile_uid,
						u.username,u.profile_picture,
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
					
					JOIN Follow f ON t.uid = f.followed_uid AND f.follower_uid = ?
					LEFT JOIN User u ON t.uid = u.uid
			        ORDER BY
			            t.created_at DESC;`, uid, uid)

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

func GetAllFollower(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodGet:
		uid := r.URL.Query().Get("uid") // To be filled
		if uid == "" {
			log.Println("fail: uid is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		rows, err := database.Db.Query(`
			        SELECT
			            u.uid,u.username,u.profile_picture
			        FROM
			            User u
			        INNER JOIN
			            Follow f ON u.uid = f.follower_uid
					where	f.followed_uid = ?;`, uid)

		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tweets := make([]models.FollowerListForHTTPGET, 0)
		for rows.Next() {
			var u models.FollowerListForHTTPGET

			if err := rows.Scan(&u.Uid, &u.Username, &u.ProfilePicture); err != nil {
				log.Printf("fail: rows.Scan, %v\n", err)

				if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
					log.Printf("fail: rows.Close(), %v\n", err)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

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

func GetAllFollowing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodGet:
		uid := r.URL.Query().Get("uid") // To be filled
		if uid == "" {
			log.Println("fail: uid is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		rows, err := database.Db.Query(`
			        SELECT
			            u.uid,u.username,u.profile_picture
			        FROM
			            User u
			        INNER JOIN
			            Follow f ON u.uid = f.followed_uid
					where	f.follower_uid = ?;`, uid)

		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tweets := make([]models.FollowerListForHTTPGET, 0)
		for rows.Next() {
			var u models.FollowerListForHTTPGET

			if err := rows.Scan(&u.Uid, &u.Username, &u.ProfilePicture); err != nil {
				log.Printf("fail: rows.Scan, %v\n", err)

				if err := rows.Close(); err != nil { // 500を返して終了するが、その前にrowsのClose処理が必要
					log.Printf("fail: rows.Close(), %v\n", err)
				}
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

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
