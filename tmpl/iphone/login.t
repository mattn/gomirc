{{define "login"}}
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
</head>
<body>
<div data-role="page" data-theme="b">
<div data-role="header" data-theme="b">
<h2>GoMIRC</h2>
</div>
<div data-role="content">
<form method="post" action="{{$root}}login/">
<input type="password" name="password"/>
<input type="submit" value="login"/>
</form>
</div>
</div>
</body>
</html>
{{end}}
