# A lite notebook use `AppendObject`

In this tutorial you will use the `AppendObject` to create a light notebook.

## Defined the fromat of notebook

```text
2015-10-25 15:20:04
this is the first note

2015-10-25 15:20:15
this is the second note


```

The notebook use a line time, and the second line note.

the time format:

```go
var currentTime = time.Now().Format("2006-01-02 15:04:05")
```

## add a line

`AppendObject` has an argument `position` is the start position to append, so first get the length of the current notebook,
then append the line text.

```go
func addLine(OSSAPI *oss.API, bucket, bookName, line string) {
	var currentTime = time.Now().Format("2006-01-02 15:04:05")
	var buf = bytes.NewBuffer(nil)
	buf.WriteString(currentTime)
	buf.WriteString("\n")
	buf.WriteString(line)
	buf.WriteString("\n\n")

	var headers = make(map[string]string)
	headers["Content-Type"] = "plain/text"

	var err error

	var headResult http.Header
	var contentLength = 0
	if headResult, err = OSSAPI.HeadObject(bucket, bookName, nil); err == nil {
		contentLength, _ = strconv.Atoi(headResult.Get("Content-Length"))
	}

	if _, err = OSSAPI.AppendObject(bucket, bookName, contentLength, buf, headers); err != nil {
		var e = oss.ParseError(err)
		log.Printf("Code: %s\nMessage: %s\n", e.Code, e.Message)
	}
}
```

## read the notebook

read the notebook is sample, just `GetObject` and print the data

```go
func readNotebook(OSSAPI *oss.API, bucket, bookName string) {
	var reader io.Reader
	var err error

	if reader, err = OSSAPI.GetObject(bucket, bookName, nil, nil); err != nil {
		var e = oss.ParseError(err)
		log.Printf("Code: %s\nMessage: %s\n", e.Code, e.Message)
	}
	var data, _ = ioutil.ReadAll(reader)
	fmt.Printf("%s\n", data)
}
```

## the notebook command line

Build a notebook command line.
`./notebook` then print the notebook content.
`./notebook 'line data'` add the line data to notebook

```go
var args = os.Args
if len(args) == 1 {
	readNotebook(OSSAPI, bucket, bookName)
} else {
	addLine(OSSAPI, bucket, bookName, args[1])
}
```

## The end

the source code [main.go](main.go)
