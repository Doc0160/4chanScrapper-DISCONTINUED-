{{define "basic"}}
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/script.js"></script>
  </head>
  <body>
    {{.Body}}
  </body>
</html>
{{end}}
{{define "basic_list"}}
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>{{.Title}}</title>
		<link rel="stylesheet" href="/static/style.css">
		<script src="/static/script.js"></script>
	</head>
	<body>
		{{.Body}}
		<ul>
		{{ range $key, $value := .List }}
			<li><a href="{{ $value.URL }}">{{ $value.Name }}</a></li>
		{{ end }}
		</ul>
	</body>
</html>
{{end}}
{{define "files"}}
<!DOCTYPE html>
<html lang="en">
	<head>
		<meta charset="utf-8">
		<title>{{.Title}}</title>
		<link rel="stylesheet" href="/static/style.css">
		<script src="/static/script.js"></script>
	</head>
	<body>
		{{.Body}}
		<br>
		<a href="{{.Prev}}">Previous</a>
		<a href="{{.Next}}">Next</a>
		<ul>
		{{ range $key, $value := .List }}
			<li><a href="{{ $value.URL }}">{{ $value.Name }}</a></li>
		{{ end }}
		</ul>
	</body>
</html>
{{end}}