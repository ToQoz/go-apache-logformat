package apachelog

import (
  "fmt"
  "io"
  "os"
  "net/http"
  "strings"
  "time"
  "regexp"
)

/*
 * import("github.com/lestrrat/go-apache-logformat-compiler")
 * l := apachelog.CombinedLog
 * l.LogLine(req)
 */

type ApacheLog struct {
  logger io.Writer
  format string
  context *replaceContext
}

type replaceContext struct {
  request     *http.Request
  status      int
  respHeader  http.Header
  reqtime     int
}

var CommonLog = NewApacheLog(
  os.Stderr,
  `%h %l %u %t "%r" %>s %b`,
)
var CombinedLog = NewApacheLog(
  os.Stderr,
  `%h %l %u %t "%r" %>s %b "%{Referer}i" "%{User-agent}i"`,
)

func NewApacheLog(w io.Writer, fmt string) *ApacheLog {
  return &ApacheLog {
    logger: w,
    format: fmt,
  }
}

func (self *ApacheLog) SetOutput(w io.Writer) {
  self.logger = w
}

/*
 * r is http.Request from client. status is the response status code.
 * respHeader is an http.Header of the response.
 *
 * reqtime is optional, and denotes the time taken to serve the
 * request in microseconds, and is optional
 *
 */
func (self *ApacheLog) LogLine(
  r           *http.Request,
  status      int,
  respHeader  http.Header,
  reqtime     int,
) {
  self.logger.Write([]byte(self.Format(r, status, respHeader, reqtime)))
}

var percentReplacer = regexp.MustCompile(
  `(?:\%\{(.+?)\}([a-zA-Z])|\%(?:[<>])?([a-zA-Z\%]))`,
)
func (self *ApacheLog) Format(
  r           *http.Request,
  status      int,
  respHeader  http.Header,
  reqtime     int,
) (string) {
  fmt := self.format
  self.context = &replaceContext {
    r,
    status,
    respHeader,
    reqtime,
  }
  return percentReplacer.ReplaceAllStringFunc(
    fmt,
    self.ReplaceFunc,
  )
}

var NilField string = "-"
func nilOrString(v string) string {
  if v == "" {
    return NilField
  } else {
    return v
  }
}

func (self *ApacheLog) ReplaceFunc (match string) string {
  r := self.context.request
  switch string(match) {
  case "%%":
    return "%"
  case "%b":
    return nilOrString(r.Header.Get("Content-Length"))
  case "%m":
    return r.Method
  case "%h":
    return nilOrString(r.RemoteAddr)
  case "%l":
    return NilField
  case "%q":
    q := r.URL.RawQuery
    if q != "" {
      return fmt.Sprintf("?%s", q)
    }
    return q
  case "%r":
    return fmt.Sprintf("%s %s %s",
      r.Method,
      r.URL,
      r.Proto,
    )
  case "%s", "%>s": // > doesn't mean anything here
    return fmt.Sprintf("%d", self.context.status)
  case "%t":
    return time.Now().Format("02/Jan/2006:15:04:05 -0700")
  case "%u":
    // Unimplemented
    return NilField
  case "%D": // custom
    if self.context.reqtime > 0 {
      return fmt.Sprintf("%d", self.context.reqtime)
    } else {
      return ""
    }
  case "%H":
    return r.Proto
  case "%T": // custom
    if self.context.reqtime > 0 {
      return fmt.Sprintf("%d", self.context.reqtime * 1000000)
    } else {
      return ""
    }
  case "%U":
    return r.URL.Path
  default:
    // if the second character isn't "{" at this point, we don't
    // know what the f this is. just return it
    if match[1] != '{' {
      return match
    }

    match = strings.TrimPrefix(match, "%{")

    var blockType byte
    // check the last character of this pattern "}i"
    for _, t := range []byte { 'i', 'o', 't' } {
      if match[len(match) - 1] == t {
        match = strings.TrimSuffix(match, fmt.Sprintf("}%c", t))
        blockType = t
        break
      }
    }

    switch blockType {
    case 'i':
      return nilOrString(r.Header.Get(match))
    case 'o':
      return nilOrString(self.context.respHeader.Get(match))
      // XXX Unimplmened
    case 't':
      // XX Unimplmented
    }
  }
  return ""
}