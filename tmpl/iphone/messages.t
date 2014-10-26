{{define "messages"}}
{{$root := .Root}}{{$value := .Value}}{{$path := .Path}}
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
<div data-role="page" data-theme="b" data-fullscreen="true" data-url="{{$path}}">
<div data-role="header" data-theme="b">
<h2>{{$value.Channel.Name | html}}@{{$value.NetworkName | html}}</h2>
</div>
<div data-role="content">
<ul data-role="listview" data-ajax="false" data-inset="true" data-theme="d">
{{range reverse $value.Channel.Messages}}{{with $message := .}}<li><span class="time">{{time $message.Time}}</span>(<span class="{{if $message.IsSelf}}self{{else}}nick{{end}}">{{nick $message.Nickname | html}}</span>)<span class="public">{{html $message.Text | clickable}}</span></li>{{end}}
{{end}}
</ul>
</div>
<div data-role="footer" data-theme="b" data-position="fixed">
<form method="post">
<input type="text" name="post"/>
<input type="submit" value="say"/>
</form>
<br />
<a accesskey="0" href="?action=refresh">refresh</a>
<a accesskey="8" href="{{$root}}">ch list</a>
</div>
</body>
</html>
{{end}}
