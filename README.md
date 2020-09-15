# GORM Oracle Driver

## Quick Start

```go
import (
	"fmt"
	oracle "github.com/cengsin/gorm-driver-oracle"
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

## 目前已知的bug

- db.Save(v interface{})在主键冲突时依旧会执行Create()而不是Update()
