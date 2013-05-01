{{define "login"}}
<html>
<head>
<meta charset="UTF-8">
<title>gomirc</title>
</head>
<body>
<h1>gomirc</h1>
<form method="post" action="{{.Root}}_login/">
<input type="password" name="password"/>
<input type="submit" value="login"/>
</form>
</body>
</html>
{{end}}
