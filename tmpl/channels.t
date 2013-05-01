{{define "channels"}}
<html>
<head>
<meta charset="UTF-8">
<title>gomirc</title>
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
{{$root := .Root}}
{{if .Value.KeywordMatches}}<a href="{{$root}}_keyword/" class="keyword">Keyword Matches!</a><br />{{end}}
{{range $channel := .Value.Channels}}<a href="{{$root}}{{urlquery $channel.NetworkName}}/{{urlquery $channel.ChannelName}}/" class="channel">{{$channel.ChannelName}}@{{$channel.NetworkName}}</a>({{new $channel.Channel}})<br />
{{end}}
<hr />
<a accesskey="0" href=".">refresh</a>
</body>
</html>
{{end}}
