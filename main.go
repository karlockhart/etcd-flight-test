package main

import (
	"context"
	"log"
	"sync"
	"time"

	"fmt"

	"github.com/coreos/etcd/clientv3"
)

var endpoints = []string{"http://localhost:2379"}
var timeout = 5 * time.Second
var wg sync.WaitGroup

func getClient() clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: timeout,
	})

	if err != nil {
		log.Fatal("Couldn't get client: ", err)
	}

	return *cli
}

func controlPoint() {
	cli := getClient()
	defer cli.Close()

	rch := cli.Watch(context.Background(), "leases", clientv3.WithPrefix())

	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}

	wg.Done()
}

func controlledProcess(name string) {
	cli := getClient()
	defer cli.Close()

	resp, err := cli.Grant(context.TODO(), 5)

	if err != nil {
		log.Fatal("Couldn't get grant.")
	}

	_, err = cli.Put(context.TODO(), fmt.Sprintf("leases-%s", name), "alive", clientv3.WithLease(resp.ID))

	if err != nil {
		log.Fatal("Couldn't put keyname")
	}
	wg.Done()
}

func main() {
	wg.Add(4)
	go controlPoint()
	go controlledProcess("test-1")
	go controlledProcess("test-2")
	go controlledProcess("test-3")
	wg.Wait()
}
