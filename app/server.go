package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

func CreateServer() ServerContext {
	return ServerContext{handlers: make(map[string]handlerFunc), handlersFuzzy: make(map[string]FuzzyParams)}
}

type handlerFunc func(RequestContext)

type ServerContext struct {
	handlers      map[string]handlerFunc
	handlersFuzzy map[string]FuzzyParams
}

type FuzzyParams struct { //TODO: expect more then one param
	paramName string
	handler   handlerFunc
}

func (ctx *ServerContext) Listen(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Failed to bind to: %s\n", addr)
		os.Exit(1)
	}

	conn, err := l.Accept()

	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	if err := ctx.AcceptConnection(conn); err != nil {
		fmt.Println(err)
		conn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
	}
}

func (ctx *ServerContext) Get(path string, handler handlerFunc) {
	if l, r, ok := strings.Cut(path, "{"); ok { //TODO: validate params
		paramName := strings.Split(r, "}")[0]
		ctx.handlersFuzzy["GET "+l] = FuzzyParams{paramName: paramName, handler: handler}
	} else {
		ctx.handlers["GET "+path] = handler
	}
}

func (ctx *ServerContext) AcceptConnection(conn net.Conn) error {
	buff := make([]byte, 1024)
	n, err := conn.Read(buff)
	if err != nil {
		return fmt.Errorf("error accepting the connection: %w", err)
	}
	reader := bufio.NewReader(bytes.NewReader(buff[:n]))
	line, _, err := reader.ReadLine()
	if err != nil {
		return fmt.Errorf("error accepting the connection: %w", err)
	}
	verb, path, _, err := ParseStartLine(line)
	if err != nil {
		return fmt.Errorf("error accepting the connection: %w", err)
	}

	if handler, params, ok := ctx.MatchPath(verb + " " + path); ok {
		rctx := RequestContextImpl{status: 200, body: make([]byte, 0), params: params, headers: make(map[string]string)}
		handler(&rctx)
		conn.Write(rctx.ToResponseBytes())
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
	return nil
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
	if h, ok := ctx.handlers[desc]; ok {
		return h, make(map[string]string), true
	} else {
		for k, v := range ctx.handlersFuzzy {
			if after, ok := strings.CutPrefix(desc, k); ok {
				val := strings.Split(after, "/")[0]
				params := make(map[string]string)
				params[v.paramName] = val
				return v.handler, params, true
			}
		}
	}
	return nil, nil, false
}

//------- Request specific logic ---------------

type RequestContext interface {
	GetParam(s string) string
	GetQuery(s string) string
	RespondWithStatusString(status int, body string)
}

type RequestContextImpl struct {
	params  map[string]string
	headers map[string]string
	body    []byte
	status  int
}

func (ctx *RequestContextImpl) GetParam(s string) string {
	return ctx.params[s]
}

func (ctx *RequestContextImpl) GetQuery(s string) string {
	return "" //TODO: implement
}

func (ctx *RequestContextImpl) RespondWithStatusString(status int, body string) {
	ctx.body = []byte(body)
	ctx.status = status
	ctx.headers["Content-Type"] = "text/plain"
	ctx.headers["Content-Length"] = strconv.Itoa(len(body))
}

func (ctx *RequestContextImpl) ToResponseBytes() []byte {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("HTTP/1.1 %v %v\r\n", ctx.status, statusCodeNames[ctx.status]))
	for k, v := range ctx.headers {
		builder.WriteString(fmt.Sprintf("%v: %v\r\n", k, v))
	}
	builder.WriteString("\r\n")
	builder.WriteString(string(ctx.body))
	return []byte(builder.String())
}

var statusCodeNames = map[int]string{
	100: "Continue",
	101: "Switching Protocols",
	102: "Processing",
	103: "Early Hints",
	200: "OK",
	201: "Created",
	202: "Accepted",
	203: "Non-Authoritative Information",
	204: "No Content",
	205: "Reset Content",
	206: "Partial Content",
	207: "Multi-Status",
	208: "Already Reported",
	226: "IM Used",
	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Found",
	303: "See Other",
	304: "Not Modified",
	307: "Temporary Redirect",
	308: "Permanent Redirect",
	400: "Bad Request",
	401: "Unauthorized",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	407: "Proxy Authentication Required",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	412: "Precondition Failed",
	413: "Content Too Large",
	414: "URI Too Long",
	415: "Unsupported Media Type",
	416: "Range Not Satisfiable",
	417: "Expectation Failed",
	418: "I'm a teapot",
	421: "Misdirected Request",
	422: "Unprocessable Content",
	423: "Locked",
	424: "Failed Dependency",
	425: "Too Early",
	426: "Upgrade Required",
	428: "Precondition Required",
	429: "Too Many Requests",
	431: "Request Header Fields Too Large",
	451: "Unavailable For Legal Reasons",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
	506: "Variant Also Negotiates",
	507: "Insufficient Storage",
	508: "Loop Detected",
	510: "Not Extended",
	511: "Network Authentication Required",
}
