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
	return ServerContext{handlers: make([]MethodDescriptor, 0), middlewares: RoutingMiddleWare}
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

	requestHeaders := make(map[string]string)
	for headerLine, _, e := reader.ReadLine(); len(headerLine) > 0 && e == nil; headerLine, _, e = reader.ReadLine() {
		hkey, hval, _ := strings.Cut(string(headerLine), ":")
		requestHeaders[hkey] = strings.Trim(hval, " ")
	}

	var requestBody []byte = make([]byte, 0)
	if cl, ok := requestHeaders["Content-Length"]; ok {
		if length, e := strconv.Atoi(cl); e == nil {
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
		responseHeaders: make(map[string]string),
		requestHeaders:  requestHeaders,
		requestBodyRaw:  requestBody,
	}
	writer := bufio.NewWriter(conn)
	ctx.middlewares(ctx, &rctx)
	WriteResponse(writer, rctx)
	writer.Flush()

	return nil
}

func RoutingMiddleWare(ctx *ServerContext, rctx *RequestContextImpl) {
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
		writer.WriteString(fmt.Sprintf("%v: %v\r\n", k, v))
	}
	writer.WriteString("\r\n")
	writer.ReadFrom(req.body)
}

//------- Request specific logic ---------------

type RequestContext interface {
	GetParam(s string) string
	GetQuery(s string) string
	GetHeader(s string) string
	RespondWithStatusString(status int, body string)
	RespondWithStatus(status int)
	GetBody() string
}

type RequestContextImpl struct {
	path            string
	method          string
	params          map[string]string
	responseHeaders map[string]string
	requestHeaders  map[string]string
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

func (ctx *RequestContextImpl) GetHeader(s string) string {
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
	ctx.responseHeaders["Content-Type"] = "text/plain"
	ctx.responseHeaders["Content-Length"] = strconv.Itoa(len(body))
}
