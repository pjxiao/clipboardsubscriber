package main

import (
    "os"
    "os/signal"
    "io/ioutil"
    "log"
    "errors"
    "github.com/urfave/cli"
    "github.com/gomodule/redigo/redis"
)



func doStuff(
    c chan<- error,
    conn redis.Conn,
    subscription string,
) {
    var (
        err error
        buffer []byte
        v interface{}
    )
    log.Printf("wait...\n")
    buffer, err = ioutil.ReadAll(os.Stdin)
    if err != nil {
        c <- err
    } else {
        body := string(buffer[:])
        v, err = conn.Do("PUBLISH", subscription, body)
        log.Printf("Published: %s", v)
        c <- err
    }
}

func prepare(
    protocol string,
    address string,
    db int,
) (redis.Conn, error) {
    return redis.Dial(
        protocol,
        address,
        redis.DialDatabase(db),
    )
}


func cleanup(conn redis.Conn) {
    log.Printf("Shutingdown\n")
    conn.Close()
    log.Printf("Shutdown\n")
}


func main() {
    app := cli.NewApp()
    app.Name = "xxx"
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name: "protocol",
            Usage: "e.g. \"tcp\"",
        },
        cli.StringFlag{
            Name: "address",
            Usage: "e.g. \"127.0.0.1:6379\"",
        },
        cli.IntFlag{
            Name: "db",
            Usage: "e.g. 1",
        },
        cli.StringFlag{
            Name: "subscription",
            Usage: "e.g. \"subscriptionname\"",
        },
    }

    app.Action = func(ctx *cli.Context) error {
        // prepare
        conn, err := prepare(
            ctx.String("protocol"),
            ctx.String("address"),
            ctx.Int("db"),
        )
        if err != nil {
            return err
        }
        defer cleanup(conn)

        // start main procedure and signal handler
        sigC := make(chan os.Signal)
        appC := make(chan error)
        signal.Notify(sigC, os.Interrupt)
        go doStuff(
            appC,
            conn,
            ctx.String("subscription"),
        )

        // wait
        select {
            case <- sigC:  // signal
                err = errors.New("Interrupted")
                log.Printf("sigC end")
            case <- appC:  // main procedure
                err = nil
                log.Printf("appC end")
        }
        return err
    }

    err := app.Run(os.Args)
    if err != nil {
        log.Fatalf("An error occured: %s", err)
    }
}

