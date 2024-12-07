package handler

import (
	"encoding/json"
	"github.com/oklog/ulid"
	"hackathon-backend/database"
	"hackathon-backend/models"
	"log"
	"net/http"
	"time"
)

func Register(w http.ResponseWriter, r *http.Request) {
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
		var req models.RegisterFromHTTPPost
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Printf("fail: json.NewDecoder.Decode, %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		//log.Printf("hogehoge %v %v\n", req.Uid, req.Email)
		/*
			if req.Name == "" || len(req.Name) > 50 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if req.Age < 20 || req.Age > 80 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

		*/
		id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()

		tx, err := database.Db.Begin()
		if err != nil {
			log.Printf("fail: db.Begin, %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		_, err = tx.Exec("INSERT INTO User (uid, email) VALUES (?, ?)", req.Uid, req.Email)
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

func RegisterUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	//w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodPost:
		var req models.RegisterUserInfoFromHTTPPost
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Printf("fail: json.NewDecoder.Decode, %v\n", err)
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

		_, err = tx.Exec("UPDATE User SET username = ?,bio = ?,profile_picture = ? where uid = ? ", req.Username, req.Bio, req.Profile, req.Uid)
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
