# gorm2sql: auto generate sql from gorm model struct


A Swiss Army Knife helps you generate sql from [gorm](https://github.com/jinzhu/gorm) model struct.


## Installation

```
go get github.com/liudanking/gorm2sql
```

## Usage

`user_email.go`:

```go
type UserBase struct {
	UserId string `sql:"index:idx_ub"`
	Ip     string `sql:"unique_index:uniq_ip"`
}

type UserEmail struct {
	Id       int64    `gorm:"primary_key"`
	UserBase
	Email      string
	Sex        bool
	Age        int
	Score      float64
	UpdateTime time.Time `sql:"default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP"`
	CreateTime time.Time `sql:"default:CURRENT_TIMESTAMP"`
}
```

```
gorm2sql sql -f user_email.go -s UserEmail -o db.sql
```

Result:

```sql
CREATE TABLE `user_email`
(
  `id` bigint AUTO_INCREMENT NOT NULL ,
  `user_id` varchar(128) NOT NULL ,
  `ip` varchar(128) NOT NULL ,
  `email` varchar(128) NOT NULL ,
  `sex` boolean NOT NULL ,
  `age` int NOT NULL ,
  `score` double NOT NULL ,
  `update_time` datetime NOT NULL  DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `create_time` datetime NOT NULL  DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_ub (`user_id`),
  UNIQUE INDEX uniq_ip (`ip`),
  PRIMARY KEY (`id`)
) engine=innodb DEFAULT charset=utf8mb4;
```


## How it works

`gorm2sql` loads go source file to golang AST, then generate sql according to `tag` of gorm struct field.
