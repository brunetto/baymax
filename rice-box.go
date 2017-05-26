package baymax

import (
	"github.com/GeertJohan/go.rice/embedded"
	"time"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "index.html.tmpl",
		FileModTime: time.Unix(1495816359, 0),
		Content:     string("<html>\n<head>\n    <meta charset=\"UTF-8\">\n    <title>{{ .Title }}</title>\n\n    <link rel=\"stylesheet\" href=\"https://maxcdn.bootstrapcdn.com/bootstrap/3.3.2/css/bootstrap.min.css\">\n    <link href='http://fonts.googleapis.com/css?family=Roboto' rel='stylesheet' type='text/css'>\n    <!-- Optional theme -->\n    <link rel=\"stylesheet\" href=\"https://maxcdn.bootstrapcdn.com/bootstrap/3.3.2/css/bootstrap-theme.min.css\">\n    <script src=\"https://ajax.googleapis.com/ajax/libs/jquery/2.1.3/jquery.min.js\"></script>\n    <!-- Latest compiled and minified JavaScript -->\n    <script src=\"https://maxcdn.bootstrapcdn.com/bootstrap/3.3.2/js/bootstrap.min.js\"></script>\n    <style type=\"text/css\">\n        body {\n            font-family: 'Roboto', sans-serif;\n        }\n        .hidden {\n            display:none;\n        }\n        pre {\n            outline: 1px solid #ccc;\n            padding: 5px; margin: 5px;\n        }\n\n        .string { color: green; }\n        .number { color: darkorange; }\n        .boolean { color: blue; }\n        .null { color: magenta; }\n        .key { color: red; }\n\n        .table     {\n            display: table;\n            width: 70%;\n            margin: 0px;\n        }\n\n        .tr        {\n            display: table-row;\n            height: 100%;\n        }\n\n        .td, .th   {\n            display: table-cell;\n            padding: 5px;\n            height: 100%;\n            max-width: 10%!important;\n        }\n        .caption   { display: table-caption; }\n\n        .th        { font-weight: bold; }\n\n        .tr:nth-child(odd) .td{\n            background-color: white;\n        }\n\n        .tr:nth-child(even) .td{\n            background-color: #eee;\n        }\n\n        span {\n            height: 100%\n        }\n\n    </style>\n\n</head>\n<body>\n<h1>{{ .Title }}</h1>\n\n<h2>Summary</h2>\n\n<div class=\"table\" style=\"width: 140px;\">\n    <div class=\"tr\">\n        <div class=\"th\" style=\"text-align: left; width: 70px;\"><span> Ok </span></div>\n        <div class=\"th\" style=\"text-align: left; width: 70px;\"><span> Warnings </span></div>\n        <div class=\"th\" style=\"text-align: left; width: 70px;\"><span> Errors </span></div>\n    </div>\n<div class=\"tr\">\n        <div class=\"td\" style=\"text-align: right; width: 70px; background-color: green\"><span> {{ .Ok }} </span></div>\n        <div class=\"td\" style=\"text-align: right; width: 70px;background-color: #ffa009\"><span> {{ .Warnings }} </span></div>\n        <div class=\"td\" style=\"text-align: right; width: 70px;background-color: red\"><span> {{ .Errors }} </span></div>\n    </div>\n</div>\n\n<h2>Details</h2>\n\n<div class=\"table\">\n    <div class=\"tr\">\n        <div class=\"th\" style=\"text-align: left;\"><span> Name </span></div>\n        <div class=\"th\" style=\"text-align: left;\"><span> URL </span></div>\n        <div class=\"th\" style=\"text-align: center;\"><span> Status </span></div>\n        <div class=\"th\" style=\"text-align: left;\"><span> Status msg </span></div>\n        <div class=\"th\" style=\"text-align: left;\"><span> Service log </span></div>\n    </div>\n    {{ range $idx, $target := .Targets }}\n    <div class=\"tr\">\n        <div class=\"td\" style=\"text-align: left;\"><span> {{ $target.Name }}</span></div>\n        <div class=\"td\" style=\"text-align: left;\"><span> <a href=\"{{ $target.URL }}\">{{ $target.URL }}</a></span></div>\n        <div class=\"td\" style=\"text-align: center;\">\n            {{ if eq $target.Status 1 }}\n            <span class=\"glyphicon glyphicon-ok\" style=\"color: green;\" aria-hidden=\"true\"></span>\n            {{ else if eq $target.Status 0 }}\n            <span class=\"glyphicon glyphicon-remove\" style=\"color: red;\" aria-hidden=\"true\"></span>\n            {{ else if eq $target.Status 2 }}\n            <span class=\"glyphicon glyphicon-bell\" style=\"color: #ffa009;\" aria-hidden=\"true\"></span>\n            {{ else }}\n            <span class=\"glyphicon glyphicon-question-sign\" aria-hidden=\"true\"></span> <span><strong>Unknown tatus: {{ $target.Status }}</strong></span>\n            {{ end }}\n        </div>\n        <div class=\"td\" style=\"text-align: left;\"><span> {{ $target.Message }}</span></div>\n        <div class=\"td\" style=\"text-align: left;\"><span> {{ $target.Logs }}</span></div>\n    </div>\n    {{ end }}\n</div>\n\n</body>\n</html>\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1495816359, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "index.html.tmpl"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`assets/templates`, &embedded.EmbeddedBox{
		Name: `assets/templates`,
		Time: time.Unix(1495816359, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"index.html.tmpl": file2,
		},
	})
}
