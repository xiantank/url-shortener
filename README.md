# url-shortener

## 設計

使用redis作為cache, 資料存入mysql中

在application層中使用redisbloom/singleflight去避免當cache失效或是不存在時打爆後面的database

短網址路徑取得方式: 使用sonyflake取得集群內的unique id, 然後 md5(uniqueID+url), 並取md5前幾byte做base62轉成短網址路徑


## 設計考量

考慮同時會有很多不同的clients訪問或嘗試使用不存在的短網址

- 多個api server可以共用的cache, 因此選用redis作為外部cache以避免db負荷太大
- 使用bloom-filter去減少不存在的短網址往db進行查詢的機會, 但仍可能發生collision, 因此繼續往後查詢的網址如果db查不到會進行cache
  - 使用[redisbloom](https://oss.redis.com/redisbloom/) 讓如果多台application server 可以直接共用同一份bloom-filter
- 使用singleflight去讓對同一個短網址的查詢不會同時重複的對db進行查詢
- TTL設定為`CACHE_TTL_IN_SECONDS` *1~1.1以減少大量cache同時到期的機會


## 該做而未做
- backup redis-bloom and restore if not exists
- redis 設定為lru的cache方式
- validate request input
- rate limit access database

## 執行步驟
假設已經有裝go, mysql

須先關掉redis, 或是在make init/run時帶如環境參數 `REDIS_PORT`


```sh
make init # run redislabs/rebloom, init database
make run # run service
```

### 可設定的環境參數

|env|default value|
|  ----  | ----  |
|SERVER_PORT|3000|
|DATABASE_HOST|localhost |
|DATABASE_PORT|3306|
|DATABASE_NAME|url_shortener|
|DATABASE_USER|root|
|DATABASE_PASSWORD|root|
|REDIS_HOST|localhost|
|REDIS_PORT|6379|
|REDIS_BLOOM_FILTER_NAME|url_shortener|
|CACHE_TTL_IN_SECONDS|86400|
