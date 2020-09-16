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


