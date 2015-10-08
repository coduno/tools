package main

import (
	"os"
	"text/template"
)

var body = `
<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{.Import}} git https://github.com/{{.Slug}}">
<meta name="go-source" content="{{.Import}} https://github.com/{{.Slug}} https://github.com/{{.Slug}}/tree/master{/dir} https://github.com/{{.Slug}}/blob/master{/dir}/{file}#L{line}">
<meta http-equiv="refresh" content="0; url=https://godoc.org/{{.Import}}">
</head>
<body>
Nothing to see here; <a href="https://godoc.org/{{.Import}}">move along</a>.
</body>
`

func main() {
	t, err := template.New("").Parse(body)
	if err != nil {
		panic(err)
	}
	t.Execute(os.Stdout, struct {
		Import, Slug string
	}{
		"cod.uno/api",
		"coduno/api",
	})
}
