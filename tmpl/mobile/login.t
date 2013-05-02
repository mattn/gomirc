{{define "login"}}
<html>
<head>
<meta charset="UTF-8">
<title>GoMIRC</title>
</head>
<body>
<h2>GoMIRC</h2>
<form method="post" action="{{.Root}}login/">
<input type="password" name="password"/>
<input type="submit" value="login"/>
</form>
</body>
</html>
{{end}}
