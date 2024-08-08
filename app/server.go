package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func CreateServer() ServerContext {
	return ServerContext{handlers: make([]MethodDescriptor, 0), middlewares: RoutingMiddleware}
}

type handlerFunc func(RequestContext)

type middlewareFuncInternal func(*ServerContext, *RequestContextImpl)
type middlewareFunc func(middlewareFuncInternal, *ServerContext, *RequestContextImpl)

type ServerContext struct {
	handlers    []MethodDescriptor
	middlewares middlewareFuncInternal
}

type MethodDescriptor struct {
	Handler handlerFunc
	Matcher Matcher
}

func (ctx *ServerContext) Listen(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to bind to: %s\n", addr)
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go ctx.AcceptConnectionAndHandleErrors(conn)
	}
}

func (ctx *ServerContext) Use(f middlewareFunc) {
	tmp := ctx.middlewares
	ctx.middlewares = func(ctx *ServerContext, rctx *RequestContextImpl) {
		f(tmp, ctx, rctx)
	}
}

func (ctx *ServerContext) Get(path string, handler handlerFunc) {
	ctx.registerHandler("GET", path, handler)
}

func (ctx *ServerContext) Post(path string, handler handlerFunc) {
	ctx.registerHandler("POST", path, handler)
}

func (ctx *ServerContext) registerHandler(verb string, path string, handler handlerFunc) error {
	if m, e := CreateUrlMatcher(verb + " " + path); e != nil {
		return e
	} else {
		ctx.handlers = append(ctx.handlers, MethodDescriptor{Matcher: m, Handler: handler})
	}
	return nil
}

func (ctx *ServerContext) AcceptConnectionAndHandleErrors(conn net.Conn) {
	if err := ctx.AcceptConnection(conn); err != nil {
		fmt.Println(err)
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
	}
	conn.Close()
}

func (ctx *ServerContext) AcceptConnection(conn net.Conn) error {
	reader := bufio.NewReader(conn)
	line, _, err := reader.ReadLine()
	if err != nil {
		return fmt.Errorf("error accepting the connection: %w", err)
	}
	verb, path, _, err := ParseStartLine(line)
	if err != nil {
		return fmt.Errorf("error accepting the connection: %w", err)
	}

	requestHeaders := make(map[string]headerValue)
	for headerLine, _, e := reader.ReadLine(); len(headerLine) > 0 && e == nil; headerLine, _, e = reader.ReadLine() {
		hName, vals := ParseHeaderLine(headerLine)
		if _, ok := requestHeaders[hName]; ok {
			requestHeaders[hName] = append(requestHeaders[hName], vals...)
		} else {
			requestHeaders[hName] = vals
		}
	}

	var requestBody []byte = make([]byte, 0)
	if cl, ok := requestHeaders["Content-Length"]; ok {
		if length, e := strconv.Atoi(cl.GetFirst()); e == nil {
			requestBody = make([]byte, length)
			io.ReadFull(reader, requestBody)
		}
	}

	rctx := RequestContextImpl{
		path:            path,
		method:          verb,
		status:          200,
		body:            strings.NewReader(""),
		params:          make(map[string]string),
		responseHeaders: make(map[string]headerValue),
		requestHeaders:  requestHeaders,
		requestBodyRaw:  requestBody,
	}
	writer := bufio.NewWriter(conn)
	ctx.middlewares(ctx, &rctx)
	WriteResponse(writer, rctx)
	writer.Flush()

	return nil
}

func RoutingMiddleware(ctx *ServerContext, rctx *RequestContextImpl) {
	if handler, params, ok := ctx.MatchPath(rctx.method + " " + rctx.path); ok {
		for k, v := range params {
			rctx.params[k] = v
		}
		handler(rctx)
	} else {
		rctx.RespondWithStatus(404)
	}
}

func ParseStartLine(l []byte) (verb string, path string, query string, err error) {
	arr := strings.Split(string(l), " ")
	if len(arr) != 3 || (arr[2] != "HTTP/1.1" && arr[2] != "HTTP/1.0") {
		return "", "", "", errors.New("not a http")
	}
	if l, r, ok := strings.Cut(arr[1], "&"); ok {
		return arr[0], l, r, nil
	} else {
		return arr[0], arr[1], "", nil
	}
}

func ParseHeaderLine(l []byte) (string, []string) {
	phase, valueStart := 0, -1

	var headerName string
	res := make([]string, 0, 1)
	for i, c := range l {
		switch {
		case valueStart <= 0 && (c == ' ' || c == '\t'):
			continue
		case phase == 0 && c == ':':
			headerName = string(l[0:i])
			phase++
		case phase == 1 && valueStart == -1:
			valueStart = i
		case phase == 1 && valueStart > 0 && c == ',':
			res = append(res, string(l[valueStart:i]))
			valueStart = -1
		}
	}

	if valueStart > 0 {
		res = append(res, string(l[valueStart:]))
	}

	return headerName, res
}

func (ctx *ServerContext) MatchPath(desc string) (handler handlerFunc, params map[string]string, ok bool) {
	for _, v := range ctx.handlers {
		if p, ok := v.Matcher(desc); ok {
			return v.Handler, p, true
		}
	}
	return nil, nil, false
}

func WriteResponse(writer *bufio.Writer, req RequestContextImpl) {
	writer.WriteString(fmt.Sprintf("HTTP/1.1 %v %v\r\n", req.status, statusCodeNames[req.status]))
	for k, v := range req.responseHeaders {
		writer.WriteString(fmt.Sprintf("%v: %v\r\n", k, strings.Join(v, ", ")))
	}
	writer.WriteString("\r\n")
	writer.ReadFrom(req.body)
}

//------- Request specific logic ---------------

type RequestContext interface {
	GetParam(s string) string
	GetQuery(s string) string
	GetHeader(s string) headerValue
	RespondWithStatusString(status int, body string)
	RespondWithStatus(status int)
	GetBody() string
}

type RequestContextImpl struct {
	path            string
	method          string
	params          map[string]string
	responseHeaders map[string]headerValue
	requestHeaders  map[string]headerValue
	body            io.Reader
	status          int
	requestBodyRaw  []byte
}

func (ctx *RequestContextImpl) GetParam(s string) string {
	return ctx.params[s]
}

func (ctx *RequestContextImpl) GetQuery(s string) string {
	return "" //TODO: implement
}

func (ctx *RequestContextImpl) GetHeader(s string) headerValue {
	return ctx.requestHeaders[s]
}

func (ctx *RequestContextImpl) GetBody() string {
	return string(ctx.requestBodyRaw)
}

func (ctx *RequestContextImpl) RespondWithStatus(status int) {
	ctx.status = status
}

func (ctx *RequestContextImpl) RespondWithStatusString(status int, body string) {
	ctx.body = strings.NewReader(body)
	ctx.status = status
	ctx.responseHeaders["Content-Type"] = []string{"text/plain"}
	ctx.responseHeaders["Content-Length"] = []string{strconv.Itoa(len(body))}
}

type headerValue []string

func (v headerValue) GetFirst() string {
	if len(v) > 0 {
		return v[0]
	} else {
		return ""
	}
}
