{{define "channels"}}
<html>
<head>
<meta charset="UTF-8">
<title>gomobirc</title>
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
{{range $nname, $network := .Value}}
{{range $cname, $channel := $network.Channels}}<a href="{{urlquery $nname}}/{{urlquery $cname}}/" class="channel">{{$cname}}@{{$nname}}</a>({{new $channel}})<br />
{{end}}
{{end}}
<hr />
<a accesskey="0" href=".">refresh</a>
</body>
</html>
{{end}}
