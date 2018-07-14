package main

import (
    "os"
    "os/signal"
    "log"
    "errors"
    "github.com/urfave/cli"
    "github.com/atotto/clipboard"
    "github.com/getlantern/systray"
    "github.com/gomodule/redigo/redis"
)



func doStuff(
    c chan<- error,
    conn redis.Conn,
    psc redis.PubSubConn,
) {
    for {
        log.Printf("wait...\n")
        switch v := psc.Receive().(type) {
            case redis.Message:
                log.Printf("%s: message %s\n", v.Channel, v.Data)
                err := clipboard.WriteAll(string(v.Data[:]))
                if err != nil {
                    c <- err
                }
                break
            case redis.Subscription:
                log.Printf("%s: %s %d \n", v.Channel, v.Kind, v.Count)
                break
            case error:
                c <- v.(error)
        }
    }
    c <- nil
}

func prepare(
    protocol string,
    address string,
    db int,
    subscription string,
) (redis.Conn, redis.PubSubConn, error) {
    var (
        conn redis.Conn
        psc redis.PubSubConn
        err error
    )
    conn, err = redis.Dial(
        protocol,
        address,
        redis.DialDatabase(db),
    )
    if err == nil {
        psc = redis.PubSubConn{Conn: conn}
        psc.Subscribe(subscription)
    }
    return conn, psc, err
}


func cleanup(conn redis.Conn, psc redis.PubSubConn) {
    log.Printf("Shutingdown\n")
    psc.Unsubscribe()
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
        conn, psc, err := prepare(
            ctx.String("protocol"),
            ctx.String("address"),
            ctx.Int("db"),
            ctx.String("subscription"),
        )
        if err != nil {
            return err
        }
        defer cleanup(conn, psc)

        // main
        c := make(chan error)
        go systray.Run(func () {
            var (
                err error
            )
            // setup systray
            systray.SetTooltip("xxx.go")

            // start main procedure and signal handler
            sigC := make(chan os.Signal)
            appC := make(chan error)
            signal.Notify(sigC, os.Interrupt)
            go doStuff(appC, conn, psc)

            // wait
            select {
                case <- sigC:  // signal
                    err = errors.New("Interrupted")
                    log.Printf("sigC end")
                case <- appC:  // main procedure
                    err = nil
                    log.Printf("appC end")
            }
            systray.Quit()
            c <- err
        }, func (){})
        return <- c
    }

    err := app.Run(os.Args)
    if err != nil {
        log.Fatalf("An error occured: %s", err)
    }
}
