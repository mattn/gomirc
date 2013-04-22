package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/fluffle/goirc/client"
	"github.com/hoisie/web"
	"github.com/mattn/go-session-manager"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var reLink = regexp.MustCompile(`(\b(https?|ftp)\://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,3}(:[a-zA-Z0-9]*)?/?([a-zA-Z0-9\-\._\?\,\'/\+&%$#\=~])*)`)

type Message struct {
	Nickname string
	Text     string
	Time     time.Time
	IsSelf   bool
	IsNotice bool
}

type Member struct {
}

type Channel struct {
	Members  map[string]*Member
	Messages []*Message
	Seen     time.Time
}

type Network struct {
	Channels map[string]*Channel
	conn     *client.Conn
	config   map[string]interface{}
}

type tmplValue struct {
	Config map[string]interface{}
	Value  interface{}
}

func timeFormat(t interface{}) string {
	return t.(time.Time).Format("15:04")
}

func nickFormat(t string) string {
	if i := strings.Index(t, "!"); i > 0 {
		t = t[:i]
	}
	return t
}

func newCount(t *Channel) int {
	n := 0
	for _, m := range t.Messages {
		if m.Time.Unix() > t.Seen.Unix() {
			n++
		}
	}
	return n
}

var configFile = flag.String("c", "config.json", "config file")

func main() {
	flag.Parse()
	f, err := os.Open(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	var config map[string]interface{}
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		log.Fatal(err)
	}

	networks := make(map[string]*Network)

	for _, elem := range config["irc"].([]interface{}) {
		irc := elem.(map[string]interface{})
		c := client.SimpleClient(irc["user"].(string), irc["user"].(string))
		c.Network = irc["name"].(string)
		c.EnableStateTracking()
		networks[c.Network] = &Network{make(map[string]*Channel), c, irc}

		c.AddHandler("connected", func(conn *client.Conn, line *client.Line) {
			joinlist := networks[conn.Network].config["channels"]
			if joinlist != nil {
				for _, ch := range joinlist.([]interface{}) {
					conn.Join(ch.(string))
				}
			}
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
			if _, ok := networks[conn.Network].Channels[line.Args[0][1:]]; !ok {
				return
			}
			networks[conn.Network].Channels[line.Args[0][1:]].Messages = append(
				networks[conn.Network].Channels[line.Args[0][1:]].Messages,
				&Message{
					line.Src,
					line.Args[1],
					time.Now(),
					nickFormat(line.Src) == networks[conn.Network].config["user"].(string),
					false,
				})
			if len(networks[conn.Network].Channels[line.Args[0][1:]].Messages) > 100 {
				networks[conn.Network].Channels[line.Args[0][1:]].Messages = networks[conn.Network].Channels[line.Args[0][1:]].Messages[1:]
			}
		})

		c.AddHandler("notice", func(conn *client.Conn, line *client.Line) {
			println("notice", line.Src, line.Args[0], line.Args[1])
			if _, ok := networks[conn.Network]; !ok {
				return
			}
			if _, ok := networks[conn.Network].Channels[line.Args[0][1:]]; !ok {
				return
			}
			networks[conn.Network].Channels[line.Args[0][1:]].Messages = append(
				networks[conn.Network].Channels[line.Args[0][1:]].Messages,
				&Message{
					line.Src,
					line.Args[1],
					time.Now(),
					nickFormat(line.Src) == networks[conn.Network].config["user"].(string),
					true,
				})
			if len(networks[conn.Network].Channels[line.Args[0][1:]].Messages) > 100 {
				networks[conn.Network].Channels[line.Args[0][1:]].Messages = networks[conn.Network].Channels[line.Args[0][1:]].Messages[1:]
			}
		})

		c.AddHandler("join", func(conn *client.Conn, line *client.Line) {
			println("join", line.Src, line.Args[0])
			if _, ok := networks[conn.Network].Channels[line.Args[0][1:]]; !ok {
				networks[conn.Network].Channels[line.Args[0][1:]] = &Channel{make(map[string]*Member), make([]*Message, 0), time.Now()}
			}
			networks[conn.Network].Channels[line.Args[0][1:]].Members[line.Src] = &Member{}
		})

		c.AddHandler("part", func(conn *client.Conn, line *client.Line) {
			println("part", line.Src, line.Args[0])
			delete(networks[conn.Network].Channels[line.Args[0][1:]].Members, line.Src)
		})

		go func(irc map[string]interface{}, c *client.Conn) {
			for {
				if err := c.Connect(irc["host"].(string), irc["password"].(string)); err != nil {
					fmt.Printf("Connection error: %s\n", err)
					return
				}
				<-quit
			}
		}(irc, c)
	}

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	manager := session.NewSessionManager(logger)
	manager.SetTimeout(10000)

	tmpl, err := template.New("gomobirc").Funcs(template.FuncMap{
		"time": timeFormat,
		"nick": nickFormat,
		"new":  newCount,
		"reverse": func(a interface{}) []interface{} {
			var b []interface{}
			ra := reflect.ValueOf(a)
			al := ra.Len()
			for n := 0; n < al; n++ {
				b = append(b, ra.Index(al-n-1).Interface())
			}
			return b
		},
		"clickable": func(s string) string {
			return reLink.ReplaceAllStringFunc(s, func(ss string) string {
				if u, err := url.Parse(ss); err == nil {
					ext := strings.ToLower(path.Ext(u.Path))
					if ext == ".jpg" || ext == ".gif" || ext == ".png" {
						return fmt.Sprintf(`<img src="http://mgw.hatena.ne.jp/?url=%s&amp;size=140x140" rel="nofollow" alt="%s"/>`, url.QueryEscape(ss), url.QueryEscape(ss))
					} else {
						return fmt.Sprintf(`<a href="http://mgw.hatena.ne.jp/?url=%s&amp;noimage=0;split=1" rel="nofollow">%s</a>`, url.QueryEscape(ss), ss)
					}
				}
				return s
			})
		},
		"eq": func(a, b string) bool {
			return a == b
		},
	}).ParseGlob(filepath.Join(filepath.Dir(os.Args[0]), "tmpl", "*.t"))
	if err != nil {
		log.Fatal(err.Error())
	}

	web.Get("/", func(ctx *web.Context) {
		if manager.GetSession(ctx, ctx.Request).Value == nil {
			ctx.Redirect(http.StatusFound, "login/")
			return
		}
		tmpl.ExecuteTemplate(ctx, "channels", tmplValue{
			Config: config,
			Value:  networks,
		})
	})

	web.Get("/login/", func(ctx *web.Context) {
		tmpl.ExecuteTemplate(ctx, "login", nil)
	})

	web.Post("/login/", func(ctx *web.Context) {
		if p, ok := ctx.Params["password"]; ok && p == config["web"].(map[string]interface{})["password"].(string) {
			manager.GetSession(ctx, ctx.Request).Value = time.Now()
			ctx.Redirect(http.StatusFound, "..")
			return
		}
		ctx.Redirect(http.StatusFound, ".")
	})

	web.Get("/(.*)/(.*)/", func(ctx *web.Context, network string, channel string) {
		if manager.GetSession(ctx, ctx.Request).Value == nil {
			ctx.Redirect(http.StatusFound, "../../login/")
			return
		}

		networks[network].Channels[channel].Seen = time.Now()
		tmpl.ExecuteTemplate(ctx, "messages", tmplValue{
			Config: config,
			Value:  networks[network].Channels[channel],
		})
	})

	web.Post("/(.*)/(.*)/", func(ctx *web.Context, network string, channel string) {
		if manager.GetSession(ctx, ctx.Request).Value == nil {
			ctx.Redirect(http.StatusFound, "../../login/")
			return
		}
		networks[network].conn.Privmsg("#"+channel, ctx.Params["post"])
		networks[network].Channels[channel].Messages = append(
			networks[network].Channels[channel].Messages,
			&Message{
				networks[network].conn.Me.Nick,
				ctx.Params["post"],
				time.Now(),
				true,
				false,
			})
		if len(networks[network].Channels[channel].Messages) > 100 {
			networks[network].Channels[channel].Messages = networks[network].Channels[channel].Messages[1:]
		}
		ctx.Redirect(http.StatusFound, ".")
	})

	web.Run(config["web"].(map[string]interface{})["addr"].(string))
}
