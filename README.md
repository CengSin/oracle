# GORM Oracle Driver

## Description

GORM Oracle driver for connect Oracle DB and Manage Oracle DB

## Required dependency Install

- see [ODPI-C Installation.](https://oracle.github.io/odpi/doc/installation.html)

## Quick Start

```go
import (
	"fmt"
	"github.com/cengsin/oracle"
	"gorm.io/gorm"
	"log"
)

func main() {
    db, err := gorm.Open(oracle.Open("system/oracle@127.0.0.1:1521/XE"), &gorm.Config{})
    if err != nil {
        // panic error or log error info
    } 
    
    // do somethings
}
```

## Unsolved Bugs

#### bug1 - Group By

by gorm [code](https://gorm.io/zh_CN/docs/query.html#Group-amp-Having): 

```go
userP := new(UserInfo)
db.Model(&UserInfo{}).Select("USER_NAME, sum(USER_AGE) as total").Where("USER_NAME like ?", "%zhang%").Group("USER_NAME").First(userP)
```

generator sql like this:

```sql
SELECT USER_NAME, sum(USER_AGE) as total
FROM USERINFO
WHERE USER_NAME like '%zhang%'
GROUP BY USER_NAME
ORDER BY USERINFO.ID FETCH NEXT 1 ROWS ONLY
```

this is a sql that have syntax errors. 

#### bug2 - TableName


If TableName() is not implemented, the default table name will become lower

```go
type Email struct {
	Id       int64  `gorm:"column:ID;primaryKey;AUTOINCREMENT"`
	EmailStr string `gorm:"column:EMAIL;NOT NULL"`
}
```

Oracle DB default name strategy is Upper String

## Easter egg 

![](https://img.shields.io/badge/-Easter%20egg-orange)

#### Create by Map

[When creating from map, hooks won’t be invoked, associations won’t be saved and primary key values won’t be back filled](https://gorm.io/zh_CN/docs/create.html#Create-From-Map)


