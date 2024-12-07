package models

import "time"

type FetchUsernameForHTTPGet struct {
	Uid            string    `json:"uid"`
	Email          string    `json:"email"`
	Username       string    `json:"username"`
	Bio            string    `json:"bio"`
	ProfilePicture string    `json:"profilePicture"`
	Created_at     time.Time `json:"created_at"`
	Follow_count   int       `json:"follow_count"`
	Followed_count int       `json:"followed_count"`
	Isfollow       bool      `json:"isfollow"`
}

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

type RegisterUserInfoFromHTTPPost struct {
	Uid      string `json:"uid"`
	Username string `json:"username"`
	Bio      string `json:"bio"`
	Profile  string `json:"profile"`
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

type LikeForHTTPPOST struct {
	Uid      string `json:"uid"`
	Tweet_id int    `json:"tweet_id"`
}

type TweetWithLikeStatus struct {
	Tweet_id       int       `json:"tweet_id"`
	Uid            string    `json:"uid"`
	Username       string    `json:"username"`
	ProfilePicture string    `json:"profilePicture"`
	Content        string    `json:"content"`
	Created_at     time.Time `json:"created_at"`
	Likes_count    int       `json:"likes_count"`
	Retweet_count  int       `json:"retweet_count"`
	IsLiked        bool      `json:"isLiked"`
}

type ReplytweetForHTTPPOST struct {
	Uid             string `json:"uid"`
	Content         string `json:"content"`
	Parent_tweet_id int    `json:"parent_tweet_id"`
}

type ReplyListForHTTPGET struct {
	Reply_id       int       `json:"reply_id"`
	Uid            string    `json:"uid"`
	Username       string    `json:"username"`
	ProfilePicture string    `json:"profilePicture"`
	Content        string    `json:"content"`
	Created_at     time.Time `json:"created_at"`
	Likes_count    int       `json:"likes_count"`
	IsLiked        bool      `json:"isLiked"`
}

type FollowForHTTPPOST struct {
	Follower_uid string `json:"follower_uid"`
	Followed_uid string `json:"followed_uid"`
}

type FollowerListForHTTPGET struct {
	Uid            string `json:"uid"`
	Username       string `json:"username"`
	ProfilePicture string `json:"profilePicture"`
}

type Gemini_Request struct {
	Prompt string `json:"prompt"`
}
