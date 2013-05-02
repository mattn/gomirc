{{define "keyword"}}
{{$root := .Root}}{{$value := .Value}}
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black">
<title>GoMIRC</title>
<link rel="stylesheet" href="{{$root}}assets/css/jquery.mobile-1.3.1.min.css" />
<script src="{{$root}}assets/javascript/jquery-1.9.1.min.js"></script>
<script src="{{$root}}assets/javascript/jquery.mobile-1.3.1.min.js"></script>
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
<div data-role="page" data-theme="b" data-fullscreen="true">
<div data-role="header" data-theme="b">
<h2>GoMIRC</h2>
</div>
<div data-role="content">
<ul data-role="listview" data-ajax="false" data-inset="true" data-theme="d">
{{range $match := $value}}<li><a href="{{$root}}irc/{{urlquery $match.NetworkName}}/{{urlquery $match.ChannelName}}/" class="channel"><span class="time">{{time $match.Message.Time}}</span>(<span class="{{if $match.Message.IsSelf}}self{{else}}nick{{end}}">{{nick $match.Message.Nickname | html}}</span>)<span class="public">{{html $match.Message.Text | clickable}}</span> ({{$match.ChannelName}}@{{$match.NetworkName}})</a></li>{{end}}
</ul>
</div>
<div data-role="footer" data-theme="b" data-position="fixed">
<a accesskey="0" href=".">refresh</a>
<a accesskey="8" href="{{$root}}">ch list</a>
</div>
</body>
</html>
{{end}}
