{{define "channels"}}
{{$root := .Root}}{{$value := .Value}}
<html>
<head>
<meta charset="UTF-8">
<title>GoMIRC</title>
<style type="text/css">
.keyword { background-color: red; }
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
<h2>GoMIRC</h2>
{{if $value.KeywordMatches}}<a href="{{$root}}keyword/" class="keyword">Keyword Matches!</a><br />{{end}}
{{range $channel := $value.Channels}}<a href="{{$root}}irc/{{urlquery $channel.NetworkName}}/{{urlquery $channel.ChannelName}}/" class="channel">{{$channel.ChannelName}}@{{$channel.NetworkName}}</a>({{new $channel.Channel}})<br />
{{end}}
<hr />
<a accesskey="0" href=".">refresh</a>
</body>
</html>
{{end}}
