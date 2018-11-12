# Transactional actions [![Build Status](https://travis-ci.org/vbereschenko/transactional.svg?branch=master)](https://travis-ci.org/vbereschenko/transactional) [![Go Report Card](https://goreportcard.com/badge/github.com/vbereschenko/transactional)](https://goreportcard.com/report/github.com/vbereschenko/transactional)

### Example of code

```go
package main

import (
	"errors"
	"github.com/vbereschenko/transactional"
	"log"
)

func main() {
	queueHandler, err := queueHandler().Build()
	if err != nil {
		log.Printf("error creating transaction: %e", err)
		return
	}
	err = queueHandler.Execute()
	log.Println("err:", err)
}

func queueHandler() transactional.Transaction {
	orderReadTransaction := transactional.Transaction{Name: "queue"}
	type Data struct {
		Id string
	}

	orderReadTransaction.Step("read", func() Data {
		return Data{"test"}
	})

	handle := func(data Data) (error) {
		if len(data.Id) > 5 {
			return errors.New("id is too big")
		}
		return nil
	}

	saveToQueue := func(err error, data Data) {}

	orderReadTransaction.FallbackStep("handle", handle, saveToQueue)

	return orderReadTransaction
}

```

#### Output will be following
```
2018/11/08 19:45:57 step [read] calling
2018/11/08 19:45:57 [queue] step [read] time 181.626µs
2018/11/08 19:45:57 step [handle] calling
2018/11/08 19:45:57 [queue] step [handle] time 13.313µs
2018/11/08 19:45:57 err: <nil>
```