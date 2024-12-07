package handler

import (
	"encoding/json"
	"hackathon-backend/database"
	"hackathon-backend/models"
	"log"
	"net/http"
)

func Like(w http.ResponseWriter, r *http.Request) {
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
		var req models.LikeForHTTPPOST
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
		err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM Likes WHERE uid = ? AND tweet_id = ?)`, req.Uid, req.Tweet_id).Scan(&exist)

		if err != nil {
			log.Printf("fail: query for like check, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if exist {
			// 既存の「いいね」を削除
			_, err := tx.Exec(`
            DELETE FROM Likes WHERE uid = ? AND tweet_id = ?`, req.Uid, req.Tweet_id)
			if err != nil {
				log.Printf("fail: delete like, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// 「いいね数」を減少
			_, err = tx.Exec(`
            UPDATE Tweet SET likes_count = likes_count - 1 WHERE tweet_id = ?`, req.Tweet_id)
			if err != nil {
				log.Printf("fail: update likes_count, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			// 新しい「いいね」を追加
			_, err := tx.Exec(`
            INSERT INTO Likes (uid, tweet_id, liked_at) VALUES (?, ?, NOW())`, req.Uid, req.Tweet_id)
			if err != nil {
				log.Printf("fail: insert like, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// 「いいね数」を増加
			_, err = tx.Exec(`
            UPDATE Tweet SET likes_count = likes_count + 1 WHERE tweet_id = ?`, req.Tweet_id)
			if err != nil {
				log.Printf("fail: update likes_count, %v\n", err)
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

func ReplyLike(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	//w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return

	case http.MethodPost:
		var req models.LikeForHTTPPOST
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
		err = tx.QueryRow(`SELECT EXISTS(SELECT 1 FROM ReplyLikes WHERE uid = ? AND reply_id = ?)`, req.Uid, req.Tweet_id).Scan(&exist)

		if err != nil {
			log.Printf("fail: query for like check, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if exist {
			// 既存の「いいね」を削除
			_, err := tx.Exec(`
            DELETE FROM ReplyLikes WHERE uid = ? AND reply_id = ?`, req.Uid, req.Tweet_id)
			if err != nil {
				log.Printf("fail: delete like, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// 「いいね数」を減少
			_, err = tx.Exec(`
            UPDATE Reply SET likes_count = likes_count - 1 WHERE reply_id = ?`, req.Tweet_id)
			if err != nil {
				log.Printf("fail: update likes_count, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		} else {
			// 新しい「いいね」を追加
			_, err := tx.Exec(`
            INSERT INTO ReplyLikes (uid, reply_id, liked_at) VALUES (?, ?, NOW())`, req.Uid, req.Tweet_id)
			if err != nil {
				log.Printf("fail: insert like, %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// 「いいね数」を増加
			_, err = tx.Exec(`
            UPDATE Reply SET likes_count = likes_count + 1 WHERE reply_id = ?`, req.Tweet_id)
			if err != nil {
				log.Printf("fail: update likes_count, %v\n", err)
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
