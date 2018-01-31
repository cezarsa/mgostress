package main

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/tsuru/config"
	"github.com/tsuru/tsuru/db"
	"github.com/tsuru/tsuru/db/storage"
	_ "gopkg.in/mgo.v2"
)

type obj struct {
	A string `bson:"_id"`
	B int
}

func test(coll *storage.Collection, i, j int) error {
	id := fmt.Sprintf("test-%d", i)
	_, err := coll.UpsertId(id, obj{A: id, B: 999})
	if err != nil {
		return err
	}
	var x obj
	err = coll.FindId(id).One(&x)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil
		}
		return err
	}
	if x.B != 999 {
		return errors.New("invalid obj")
	}
	err = coll.RemoveId(id)
	if err == mgo.ErrNotFound {
		return nil
	}
	go func() {
		time.Sleep(time.Second)
		coll.Close()
	}()
	return err
}

func main() {
	runtime.GOMAXPROCS(10)
	nGoroutines := 500
	config.Set("database:url", "127.0.0.1:27017")
	config.Set("database:name", "tsuru_mongodb_stress_test")
	stor, _ := db.Conn()
	err := stor.Apps().Database.DropDatabase()
	if err != nil {
		panic(err)
	}
	stor.Close()
	wg := sync.WaitGroup{}
	for g := 0; g < nGoroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				stor, err := db.Conn()
				if err != nil {
					panic(err)
				}
				err = test(stor.Apps(), i, g)
				if err != nil {
					panic(err)
				}
			}
		}(g)
	}
	wg.Wait()
}
