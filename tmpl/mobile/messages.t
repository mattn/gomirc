{{define "messages"}}
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
<h4>{{$value.ChannelName | html}}@{{$value.NetworkName | html}}</h4>
<form method="post">
<input type="text" name="post"/>
<input type="submit" value="say"/>
</form>
{{range reverse $value.Channel.Messages}}{{with $message := .}}<span class="time">{{time $message.Time}}</span>(<span class="{{if $message.IsSelf}}self{{else}}nick{{end}}">{{nick $message.Nickname | html}}</span>)<span class="public">{{html $message.Text | clickable}}</span><br />{{end}}
{{end}}
<hr />
<a accesskey="0" href=".">refresh</a>
<a accesskey="8" href="{{$root}}">ch list</a>
</body>
</html>
{{end}}
