package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mattn/go-mobileagent"
	"github.com/mattn/go-session-manager"
	"github.com/thoj/go-ircevent"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

var reLink = regexp.MustCompile(`(\b(https?|ftp)\://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,3}(:[a-zA-Z0-9]*)?/?([a-zA-Z0-9\-\._\?\,\'/\+&%$#\=~:!;])*)`)

var mutex sync.Mutex

type Message struct {
	Nickname string    `json:"nickname"`
	Text     string    `json:"text"`
	Time     time.Time `json:"time"`
	IsSelf   bool      `json:"is_self"`
	IsNotice bool      `json:"is_notice"`
}

type Member struct {
	IsOwner bool `json:"is_owner"`
}

type Channel struct {
	Name     string             `json:"name"`
	Members  map[string]*Member `json:"members"`
	Messages []*Message         `json:"messages"`
	Seen     time.Time          `json:"seen"`
}

type ChannelMap struct {
	NetworkName string
	ChannelName string
	Channel     *Channel
}

type Channels []*ChannelMap

func (p Channels) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Channels) Len() int           { return len(p) }
func (p Channels) Less(i, j int) bool { return newCount(p[i].Channel) >= newCount(p[j].Channel) }

type KeywordMatch struct {
	NetworkName string
	ChannelName string
	Message     *Message
}

type Network struct {
	Channels map[string]*Channel `json:"channels"`
	conn     *irc.Connection
	config   map[string]interface{}
}

type tmplValue struct {
	Root  interface{}
	Path  interface{}
	Value interface{}
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

func getChannelName(name string) string {
	if name != "" && name[0] == '#' {
		return name[1:]
	}
	return name
}

func ircLowerCaseMap(r rune) rune {
	switch {
	case r >= 'A' && r <= 'Z':
		return 'a' + (r - 'A')
	case r == '[':
		return '{'
	case r == ']':
		return '}'
	case r == '\\':
		return '|'
	}
	return r
}

func getChannel(network *Network, channel string) *Channel {
	ircChannelName := strings.Map(ircLowerCaseMap, channel)
	if _, ok := network.Channels[ircChannelName]; !ok {
		network.Channels[ircChannelName] = &Channel{channel, make(map[string]*Member), make([]*Message, 0), time.Now()}
	}
	return network.Channels[ircChannelName]
}

func getTmplName(req *http.Request) string {
	userAgent := req.Header.Get("User-Agent")
	if mobileagent.IsMobile(userAgent) {
		return "mobile"
	}
	return "iphone"
}

func weblog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

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
	f.Close()

	networks := make(map[string]*Network)
	if backlog, ok := config["web"].(map[string]interface{})["backlog"].(string); ok {
		if f, err = os.Open(backlog); err == nil {
			json.NewDecoder(f).Decode(&networks)
			f.Close()
		}
		sc := make(chan os.Signal)
		signal.Notify(sc, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
		go func() {
			<-sc
			if f, err = os.Create(backlog); err == nil {
				json.NewEncoder(f).Encode(&networks)
				f.Close()
			}
			os.Exit(0)
		}()
	}

	for _, nw := range networks {
		for cn, cc := range nw.Channels {
			if len(cc.Name) == 0 {
				cc.Name = cn
				ircChannelName := strings.Map(ircLowerCaseMap, cn)
				nw.Channels[ircChannelName] = cc
				delete(nw.Channels, cn)
			}
		}
	}

	keywords := []string{}
	if kwi, ok := config["web"].(map[string]interface{})["keywords"].([]interface{}); ok {
		for _, kw := range kwi {
			keywords = append(keywords, kw.(string))
		}
	}

	keywordMatches := []*KeywordMatch{}

	nameOf := func(c *irc.Connection) string {
		for k, v := range networks {
			if v.conn == c {
				return k
			}
		}
		panic("Shouldn't pass")
	}

	for _, elem := range config["irc"].([]interface{}) {
		cfg := elem.(map[string]interface{})
		c := irc.IRC(cfg["nick"].(string), cfg["user"].(string))
		if network, ok := networks[cfg["name"].(string)]; ok {
			network.conn = c
			network.config = cfg
		} else {
			networks[cfg["name"].(string)] = &Network{make(map[string]*Channel), c, cfg}
		}

		c.AddCallback("CONNECTED", func(e *irc.Event) {
			mutex.Lock()
			defer mutex.Unlock()
			joinlist := networks[nameOf(e.Connection)].config["channels"]
			if joinlist != nil {
				for _, ch := range joinlist.([]interface{}) {
					e.Connection.Join(ch.(string))
				}
			}
		})

		quit := make(chan bool)
		c.AddCallback("DISCONNECTED", func(e *irc.Event) {
			quit <- true
		})

		c.AddCallback("PRIVMSG", func(e *irc.Event) {
			mutex.Lock()
			defer mutex.Unlock()
			println("PRIVMSG", e.Source, e.Arguments[0], e.Arguments[1])
			if _, ok := networks[nameOf(e.Connection)]; !ok {
				return
			}
			message := &Message{
				e.Source,
				e.Arguments[1],
				time.Now(),
				nickFormat(e.Source) == networks[nameOf(e.Connection)].config["nick"].(string),
				false,
			}
			ch := getChannel(networks[nameOf(e.Connection)], getChannelName(e.Arguments[0]))
			ch.Messages = append(ch.Messages, message)
			if len(ch.Messages) > 100 {
				ch.Messages = ch.Messages[1:]
			}
			for _, keyword := range keywords {
				if strings.Contains(e.Arguments[1], keyword) {
					keywordMatches = append(keywordMatches, &KeywordMatch{nameOf(e.Connection), getChannelName(e.Arguments[0]), message})
				}
			}
		})

		c.AddCallback("notice", func(e *irc.Event) {
			mutex.Lock()
			defer mutex.Unlock()
			println("NOTICE", e.Source, e.Arguments[0], e.Arguments[1])
			if _, ok := networks[nameOf(e.Connection)]; !ok {
				return
			}
			message := &Message{
				e.Source,
				e.Arguments[1],
				time.Now(),
				nickFormat(e.Source) == networks[nameOf(e.Connection)].config["nick"].(string),
				true,
			}
			ch := getChannel(networks[nameOf(e.Connection)], getChannelName(e.Arguments[0]))
			ch.Messages = append(ch.Messages, message)
			if len(ch.Messages) > 100 {
				ch.Messages = ch.Messages[1:]
			}
			for _, keyword := range keywords {
				if strings.Contains(e.Arguments[1], keyword) {
					keywordMatches = append(keywordMatches, &KeywordMatch{nameOf(e.Connection), getChannelName(e.Arguments[0]), message})
				}
			}
		})

		c.AddCallback("JOIN", func(e *irc.Event) {
			mutex.Lock()
			defer mutex.Unlock()
			println("JOIN", e.Source, e.Arguments[0])
			if _, ok := networks[nameOf(e.Connection)]; !ok {
				return
			}
			members := getChannel(networks[nameOf(e.Connection)], getChannelName(e.Arguments[0])).Members
			members[e.Source] = &Member{}
		})

		c.AddCallback("PART", func(e *irc.Event) {
			println("PART", e.Source, e.Arguments[0])
			members := getChannel(networks[nameOf(e.Connection)], getChannelName(e.Arguments[0])).Members
			delete(members, e.Source)
		})

		c.Password = cfg["password"].(string)
		go func(cfg map[string]interface{}, c *irc.Connection) {
			for {
				if err := c.Connect(cfg["host"].(string)); err != nil {
					fmt.Printf("Connection error: %s\n", err)
					return
				}
				<-quit
			}
		}(cfg, c)
	}

	manager := session.NewSessionManager(nil)
	manager.SetTimeout(60 * 60 * 24)
	root := "/"
	if root, _ = config["web"].(map[string]interface{})["root"].(string); root != "/" {
		manager.SetPath(root)
	}
	if !strings.HasSuffix(root, "/") {
		root += "/"
		config["web"].(map[string]interface{})["root"] = root
	}

	fmap := template.FuncMap{
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
						return fmt.Sprintf(`<a href="%s"><img src="%s" rel="nofollow" alt="%s"/></a>`, ss, ss, url.QueryEscape(ss))
					} else {
						return fmt.Sprintf(`<a href="%s" rel="nofollow">%s</a>`, ss, ss)
					}
				}
				return s
			})
		},
		"clickable_mobile": func(s string) string {
			return reLink.ReplaceAllStringFunc(s, func(ss string) string {
				if u, err := url.Parse(ss); err == nil {
					ext := strings.ToLower(path.Ext(u.Path))
					if ext == ".jpg" || ext == ".gif" || ext == ".png" {
						return fmt.Sprintf(`<a href="http://mgw.hatena.ne.jp/?url=%s&"><img src="http://mgw.hatena.ne.jp/?url=%s&amp;size=140x140" rel="nofollow" alt="%s"/></a>`, url.QueryEscape(ss), url.QueryEscape(ss), url.QueryEscape(ss))
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
	}

	tmpls := map[string]*template.Template{}
	rootdir := filepath.Dir(os.Args[0])
	if tmpldir, ok := config["web"].(map[string]interface{})["rootdir"].(string); ok {
		rootdir = tmpldir
	}
	logdir, _ := config["web"].(map[string]interface{})["logdir"].(string)

	tmpls["mobile"], err = template.New("mobile").Funcs(fmap).ParseGlob(filepath.Join(rootdir, "tmpl/mobile", "*.t"))
	if err != nil {
		log.Fatal("mobile ", err.Error())
	}
	tmpls["iphone"], err = template.New("iphone").Funcs(fmap).ParseGlob(filepath.Join(rootdir, "tmpl/iphone", "*.t"))
	if err != nil {
		log.Fatal("iphone ", err.Error())
	}

	http.HandleFunc(root+"assets/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(rootdir, "static/"+r.URL.Path[len(root+"asserts"):]))
	})

	if len(logdir) > 0 {
		http.Handle(root+"log/", http.StripPrefix(root+"log/", http.FileServer(http.Dir(logdir))))
	}

	http.HandleFunc(root, func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		if manager.GetSession(w, r).Value == nil {
			http.Redirect(w, r, root+"login/", http.StatusFound)
			return
		}
		chs := make(Channels, 0)
		for nn, nw := range networks {
			for cn, cc := range nw.Channels {
				chs = append(chs, &ChannelMap{
					NetworkName: nn,
					ChannelName: cn,
					Channel:     cc,
				})
			}
		}
		sort.Sort(chs)
		w.Header().Add("Cache-Control", "max-age=0")
		tmpls[getTmplName(r)].ExecuteTemplate(w, "channels", tmplValue{
			Root: root,
			Path: r.URL.Path,
			Value: &struct {
				Channels       Channels
				KeywordMatches []*KeywordMatch
				HasLog         bool
			}{
				Channels:       chs,
				KeywordMatches: keywordMatches,
				HasLog:         len(logdir) > 0,
			},
		})
	})

	http.HandleFunc(root+"login/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Add("Cache-Control", "max-age=0")
			tmpls[getTmplName(r)].ExecuteTemplate(w, "login", tmplValue{
				Root:  root,
				Path:  r.URL.Path,
				Value: nil,
			})
		case "POST":
			if p := r.FormValue("password"); p == config["web"].(map[string]interface{})["password"].(string) {
				manager.GetSession(w, r).Value = time.Now()
				http.Redirect(w, r, root, http.StatusFound)
				return
			}
			http.Redirect(w, r, ".", http.StatusFound)
		}
	})

	http.HandleFunc(root+"keyword/", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		if manager.GetSession(w, r).Value == nil {
			http.Redirect(w, r, root+"login/", http.StatusFound)
			return
		}

		w.Header().Add("Cache-Control", "max-age=0")
		tmpls[getTmplName(r)].ExecuteTemplate(w, "keyword", tmplValue{
			Root:  root,
			Path:  r.URL.Path,
			Value: keywordMatches,
		})
		keywordMatches = []*KeywordMatch{}
	})

	http.HandleFunc(root+"irc/", func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		if manager.GetSession(w, r).Value == nil {
			http.Redirect(w, r, root+"login/", http.StatusFound)
			return
		}

		paths := strings.Split(r.URL.Path[len(root+"irc/"):], "/")
		network, channel := paths[0], paths[1]

		switch r.Method {
		case "GET":
			ch := getChannel(networks[network], channel)
			ch.Seen = time.Now()
			w.Header().Add("Cache-Control", "max-age=0")
			tmpls[getTmplName(r)].ExecuteTemplate(w, "messages", tmplValue{
				Root: root,
				Path: r.URL.Path,
				Value: &ChannelMap{
					NetworkName: network,
					ChannelName: channel,
					Channel:     ch,
				},
			})
		case "POST":
			p := r.FormValue("post")
			if p != "" {
				if p[0] == '/' {
					networks[network].conn.SendRaw(p[1:])
				} else {
					networks[network].conn.Privmsg("#"+channel, p)
					ch := getChannel(networks[network], channel)
					ch.Seen = time.Now()
					nick := networks[network].config["nick"].(string)
					ch.Messages = append(
						ch.Messages,
						&Message{
							nick,
							r.FormValue("post"),
							time.Now(),
							true,
							false,
						})
					if len(ch.Messages) > 100 {
						ch.Messages = ch.Messages[1:]
					}
				}
			}
			http.Redirect(w, r, r.URL.Path, http.StatusFound)
		}
	})

	addr := config["web"].(map[string]interface{})["addr"].(string)
	typ, _ := config["web"].(map[string]interface{})["type"]
	typs, _ := typ.(string)
	if typs != "" {
		l, err := net.Listen(typs, addr)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal(http.Serve(l, http.DefaultServeMux))
	} else {
		log.Fatal(http.ListenAndServe(addr, weblog(http.DefaultServeMux)))
	}
}
