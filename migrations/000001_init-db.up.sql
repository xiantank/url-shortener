CREATE Table `short_urls` (
    id int primary key auto_increment not null,
    uid varchar(10) unique not null,
    url varchar(500) default '' not null,
    expire_at  int unsigned not null
    -- TODO: should use index for expires_at ?
)DEFAULT CHARSET=utf8mb4