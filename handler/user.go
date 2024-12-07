package handler

import (
	"encoding/json"
	"github.com/oklog/ulid"
	"hackathon-backend/database"
	"hackathon-backend/models"
	"log"
	"math/rand"
	"net/http"
	"time"
)

var entropy = ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)

// ② /userでリクエストされたらnameパラメーターと一致する名前を持つレコードをJSON形式で返す
func Handler(w http.ResponseWriter, r *http.Request) {
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
		var req models.UserInputFromHTTPPost
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Printf("fail: json.NewDecoder.Decode, %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if req.Name == "" || len(req.Name) > 50 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.Age < 20 || req.Age > 80 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()

		tx, err := database.Db.Begin()
		if err != nil {
			log.Printf("fail: db.Begin, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec("INSERT INTO user (id, name, age) VALUES (?, ?, ?)", id, req.Name, req.Age)
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

		resp := map[string]string{"id": id}
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

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodGet:
		rows, err := database.Db.Query("SELECT id, name, age FROM user")
		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

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

		bytes, err := json.Marshal(users)
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

func FetchUsername(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodGet:

		uid := r.URL.Query().Get("uid")
		login_uid := r.URL.Query().Get("login")

		if uid == "" {
			log.Println("fail: uid is empty")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		rows, err := database.Db.Query(`

		SELECT 
            t.uid,t.email, t.username, t.bio,t.profile_picture,t.created_at,
			t.follow_count,t.follower_count,
            CASE WHEN l.follower_uid IS NOT NULL  THEN TRUE ELSE FALSE END AS liked_by_user
        FROM 
            User t
        LEFT JOIN 
            Follow l
        ON 
            l.follower_uid = ? AND l.followed_uid = t.uid
        WHERE 
            t.uid = ?`, login_uid, uid)

		if err != nil {
			log.Printf("fail: db.Query, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		tweets := make([]models.FetchUsernameForHTTPGet, 0)
		for rows.Next() {
			var u models.FetchUsernameForHTTPGet
			var createdAt []byte // まずバイト列で受け取る
			if err := rows.Scan(&u.Uid, &u.Email, &u.Username, &u.Bio, &u.ProfilePicture, &createdAt, &u.Follow_count, &u.Followed_count, &u.Isfollow); err != nil {
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
