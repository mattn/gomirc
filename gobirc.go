package main

import (
	"fmt"
	"github.com/fluffle/goirc/client"
	//"github.com/fluffle/goirc/state"
	"github.com/hoisie/web"
	//"net/http"
	"strings"
	"time"
	"text/template"
)

type Message struct {
	Nickname string
	Text string
	Time string
	Notice bool
}

type Member struct {
}

type Channel struct {
	Members  map[string]*Member
	Messages []*Message
}

type Network struct {
	Channels map[string]*Channel
}

func simpleNick(nickname string) string {
	if i := strings.Index(nickname, "!"); i > 1 {
		nickname = nickname[:i]
	}
	return nickname
}

func main() {
	networks := make(map[string]*Network)

	c := client.SimpleClient("lingr", "lingr")
	c.Network = c.Me.Ident
	c.EnableStateTracking()

	c.AddHandler("connected", func(conn *client.Conn, line *client.Line) {
		networks[conn.Network] = &Network { make(map[string]*Channel) }
	})

	quit := make(chan bool)
	c.AddHandler("disconnected", func(conn *client.Conn, line *client.Line) {
		quit <- true
	})

	c.AddHandler("privmsg", func(conn *client.Conn, line *client.Line) {
		println("privmsg", line.Src, line.Args[0], line.Args[1])
		if _, ok := networks[conn.Network]; !ok {
			return
		}
		if _, ok := networks[conn.Network].Channels[line.Args[0]]; !ok {
			return
		}
		networks[conn.Network].Channels[line.Args[0]].Messages = append(
			networks[conn.Network].Channels[line.Args[0]].Messages,
			&Message {
				line.Src,
				line.Args[1],
				time.Now().Format("15:04"),
				false,
			})
		if len(networks[conn.Network].Channels[line.Args[0]].Messages) > 100 {
			networks[conn.Network].Channels[line.Args[0]].Messages = networks[conn.Network].Channels[line.Args[0]].Messages[1:]
		}
	})

	c.AddHandler("notice", func(conn *client.Conn, line *client.Line) {
		println("notice", line.Src, line.Args[0], line.Args[1])
		if _, ok := networks[conn.Network]; !ok {
			return
		}
		if _, ok := networks[conn.Network].Channels[line.Args[0]]; !ok {
			return
		}
		networks[conn.Network].Channels[line.Args[0]].Messages = append(
			networks[conn.Network].Channels[line.Args[0]].Messages,
			&Message {
				line.Src,
				line.Args[1],
				time.Now().Format("15:04"),
				true,
			})
		if len(networks[conn.Network].Channels[line.Args[0]].Messages) > 100 {
			networks[conn.Network].Channels[line.Args[0]].Messages = networks[conn.Network].Channels[line.Args[0]].Messages[1:]
		}
	})

	c.AddHandler("join", func(conn *client.Conn, line *client.Line) {
		println("join", line.Src, line.Args[0])
		if _, ok := networks[conn.Network].Channels[line.Args[0]]; !ok {
			networks[conn.Network].Channels[line.Args[0]] = &Channel { make(map[string]*Member), make([]*Message, 0) }
		}
		networks[conn.Network].Channels[line.Args[0]].Members[line.Src] = &Member {}
	})

	c.AddHandler("part", func(conn *client.Conn, line *client.Line) {
		println("part", line.Src, line.Args[0])
		delete(networks[conn.Network].Channels[line.Args[0]].Members, line.Src)
	})

	web.Get("/", func(ctx *web.Context) {
		tpl, err := template.ParseFiles("channels.tmpl")
		if err != nil {
			ctx.Abort(500, err.Error())
		}
		tpl.Execute(ctx, networks)
	})

	web.Get("/(.*)/(.*)", func(ctx *web.Context, network string, channel string) {
		tpl, err := template.ParseFiles("messages.tmpl")
		if err != nil {
			ctx.Abort(500, err.Error())
		}
		tpl.Execute(ctx, networks[network].Channels[channel])
	})

	go func() {
		web.Run(":5004")
	}()

	for {
		if err := c.Connect("localhost:6668", "hogemoge"); err != nil {
			fmt.Printf("Connection error: %s\n", err)
			return
		}
		<-quit
	}
}
