# GORM Oracle Driver

![](https://starchart.cc/CengSin/oracle.svg)

## Description

GORM Oracle driver for connect Oracle DB and Manage Oracle DB, Based on [stevefan1999-personal/gorm-driver-oracle](https://github.com/stevefan1999-personal/gorm-driver-oracle)
，not recommended for use in a production environment

## Required dependency Install

- Oracle 12C+
- Golang 1.13+
- see [ODPI-C Installation.](https://oracle.github.io/odpi/doc/installation.html)
- gorm 1.23.4+

## Quick Start
### how to install 
```bash
go get github.com/dzwvip/oracle
```
###  usage

```go
import (
	"fmt"
	"github.com/dzwvip/oracle"
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
