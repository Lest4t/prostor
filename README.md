## GoDevino library

Prostor library for api site http://prostor-sms.ru

###Installation:

```bash
go get github.com/Lest4t/prostor
```

### Examples

```go
package main

import (
    "fmt"
    "github.com/Lest4t/prostor"
)

func main() {
    var sms *prostor.Client

    prostor.Username = "user"
    prostor.Password = "password"

    balance, err := sms.GetBalance()
    if err != nil {
        fmt.Printf("[ERR] %s\n", err)
    } else {
        fmt.Printf("[INF] Balance: %s\n", balance)
    }
    
    message_id, err := sms.SendMessage("", "+79320001112", "Happy New Year", "")
    if err != nil {
        fmt.Printf("[ERR] %s\n", err)
    } else {
        fmt.Printf("[INF] Message sended.\n")
    }

    if message_id != "" {
        status, err := sms.GetMessageState(message_id)
        if err != nil {
        fmt.Printf("[ERR] %s\n", err)
    } else {
        fmt.Printf("[INF] status: %s\n", status)
    }
}
```
