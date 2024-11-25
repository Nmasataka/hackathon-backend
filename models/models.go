package models

import "time"

type UserResForHTTPGet struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type UserInputFromHTTPPost struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type RegisterFromHTTPPost struct {
	Uid   string `json:"uid"`
	Email string `json:"email"`
}

type LoginForHTTPGET struct {
	Email string `json:"email"`
}

type TweetForHTTPPOST struct {
	Uid     string `json:"uid"`
	Content string `json:"content"`
}

type TweetListForHTTPGET struct {
	Tweet_id   string    `json:"tweet_id"`
	Uid        string    `json:"uid"`
	Content    string    `json:"content"`
	Created_at time.Time `json:"created_at"`
	//Updated_at    time.Time `json:"updated_at"`
	Likes_count int `json:"likes_count"`
	//Retweet_count int `json:"retweet_count"`
}
