# hackathon-backend

Go + MySQL で動いているバックエンドAPIサーバー(ポート8000で待ち受け)。

## ローカルでの動かし方

### 1. MySQLをDockerで起動する

MySQLコンテナを3306番ポートで起動する。

### 2. `database/db.go` の接続設定を確認する

`database/db.go` の `InitDB()` には接続先の書き方が2パターン用意されている。

- 上のコメントアウトされているブロック(`mysqlHost` を使う版) → Docker Compose等で`MYSQL_HOST`にサービス名を渡す場合用
- その下の有効になっているブロック(`godotenv.Load` 〜 `localhost:3306` に接続する版) → **ローカル実行用**

ローカルで動かす場合は、下の`localhost:3306`に接続するブロックが有効(コメントアウトされていない)状態になっていることを確認する。上のブロックがコメントインされていたらコメントアウトし、下のブロックがコメントアウトされていたら解除する。

なお、ローカル版をコメントインすると、`"github.com/joho/godotenv"`が自動でimportされる。環境変数を読むためである。

### 3. 依存関係を取得する

```bash
go mod tidy
```

### 4. サーバーを起動する

```bash
go run main.go
```

`Listening...` と表示されればMySQLへの接続に成功し、8000番ポートで待ち受けている状態。

### 5. curlで動作確認する

```bash
# 全ユーザー取得
curl http://localhost:8000/users

# クエリパラメータ付き(name=hanakoに一致するユーザーを取得)
# ※ zshでは ? や & がglobとして解釈されるため、URLは必ずダブルクォートで囲む
curl "http://localhost:8000/user?name=hanako"
```

JSON形式でレスポンスが返ってくれば正常に動作している。
