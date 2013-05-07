{{define "keyword"}}
{{$root := .Root}}{{$value := .Value}}
<html>
<head>
<meta charset="UTF-8">
<title>GoMIRC</title>
<style type="text/css">
.keyword_recent_notice { background-color: red; }
.time { color: #004080; }
.channel { color: #004080; }
.notice { color: #808080; }
.join { color: #ccfece; }
.leave { color: #cccefe; }
.ctcp_action { color: #808080; font-style: italic; }
.kick { color: #fc4efe; }
.snotice { color: #408080; }
.connect { color: #408080; }
.reconnect { color: #408080; }
.nick { color: #000080; }
.self { color: #6060a0; }
</style>
</head>
<body>
<h4>GoMIRC</h4>
{{range $match := $value}}<span class="time">{{time $match.Message.Time}}</span>(<span class="{{if $match.Message.IsSelf}}self{{else}}nick{{end}}">{{nick $match.Message.Nickname | html}}</span>)<span class="public">{{html $match.Message.Text | clickable}}</span> (<a href="{{$root}}irc/{{urlquery $match.NetworkName}}/{{urlquery $match.ChannelName}}/" class="channel">{{$match.ChannelName}}@{{$match.NetworkName}}</a>)<br />{{end}}
<hr />
<a accesskey="0" href=".">refresh</a>
<a accesskey="8" href="{{$root}}">ch list</a>
</body>
</html>
{{end}}
