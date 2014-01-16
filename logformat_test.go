package apachelog

import (
  "fmt"
  "net/http"
  "os"
  "testing"
  "time"
)

func TestBasic(t *testing.T) {
  l := CombinedLog
  r, err := http.NewRequest("GET", "http://golang.org", nil)
  if err != nil {
    t.Errorf("Failed to create request: %s", err)
  }
  r.RemoteAddr = "127.0.0.1"
  r.Header.Set("User-Agent", "Apache-LogFormat Port In Golang")
  r.Header.Set("Referer", "http://dummy.com")
  output := l.Format(
    r,
    200,
    http.Header{ "Content-Type": []string{"text/plain"} },
    1500000,
  )
  if output == "" {
    t.Errorf("Failed to Format")
  }
  t.Logf(`output = "%s"`, output)
}

func TestAllTypes(t *testing.T) {
  l := NewApacheLog(os.Stderr, "This should be a verbatim percent sign -> %%")
  output := l.Format(
    &http.Request{},
    200,
    http.Header{},
    0,
  )

  if output != "This should be a verbatim percent sign -> %" {
    t.Errorf("Failed to format. Got '%s'", output)
  }
}

func TestResponseHeader(t *testing.T) {
  l := NewApacheLog(os.Stderr, "%{X-Req-Header}i %{X-Resp-Header}o")
  r, err := http.NewRequest("GET", "http://golang.org", nil)
  if err != nil {
    t.Errorf("Failed to create request: %s", err)
  }

  r.Header.Set("X-Req-Header", "Gimme a response!")

  output := l.Format(r, 200, http.Header{"X-Resp-Header": []string{"Here's your response"}}, 1000000)
  if output != "Gimme a response! Here's your response" {
    t.Errorf("output '%s' did not match", output)
  }
  t.Logf("%s", output)
}

func TestQuery(t *testing.T) {
  l := NewApacheLog(os.Stderr, "%m %U %q %H")
  r, err := http.NewRequest("GET", "http://golang.org/foo?bar=baz", nil)
  if err != nil {
    t.Errorf("Failed to create request: %s", err)
  }

  output := l.Format(r, 200, http.Header{}, 1000000)
  if output != "GET /foo ?bar=baz HTTP/1.1" {
    t.Errorf("output '%s' did not match", output)
  }
  t.Logf("%s", output)
}

func TestElpasedTime (t *testing.T) {
  l := NewApacheLog(os.Stderr, "%T %D")
  output := l.Format(&http.Request{}, 200, http.Header{}, 1 * time.Second)
  if output != "1 1000000" {
    t.Errorf("output '%s' did not match", output)
  }
  t.Logf("%s", output)
}

func TestClone(t *testing.T) {
  l := CombinedLog.Clone()
  l.SetOutput(os.Stdout)

  if CombinedLog.logger == l.logger {
    t.Errorf("logger struct must not be the same")
  }
}

func TestEdgeCase(t *testing.T) {
  // stray %
  l := NewApacheLog(os.Stderr, "stray percent at the end: %")
  output := l.Format(
    &http.Request {},
    200,
    http.Header {},
    0,
  )
  if output != "stray percent at the end: %" {
    t.Errorf("Failed to match output")
    t.Logf("Expected '%s', got '%s'", "stray percent at the end %", output)
  }

  // %{...} with missing }
  l = NewApacheLog(os.Stderr, "Missing closing brace: %{Test <- this should be verbatim")
  r, _ := http.NewRequest("GET", "http://golang.com", nil)
  r.Header.Set("Test", "Test Me Test Me")
  output = l.Format(
    r,
    200,
    http.Header {},
    0,
  )
  if output != "Missing closing brace: %{Test <- this should be verbatim" {
    t.Errorf("Failed to match output")
    t.Logf("Exepected '%s', got '%s'",
      "Missing closing brace: %{Test <- this should be verbatim",
      output,
    )
  }

  // %s and %>s should be the same in our case
  l = NewApacheLog(os.Stderr, "%s = %>s")
  output = l.Format(
    r,
    404,
    http.Header {},
    0,
  )
  if output != "404 = 404" {
    t.Errorf("%%s and %%>s should be the same. Expected '404 = 404', got '%s'", output)
  }

  // pid
  l = NewApacheLog(os.Stderr, "%p")
  output = l.Format(
    r,
    200,
    http.Header {},
    0,
  )
  if output != fmt.Sprintf("%d", os.Getpid()) {
    t.Errorf("%%p should get us our own pid. Expected '%d', got '%s'", os.Getpid(), output)
  }
}

func BenchmarkReplaceLoop(t *testing.B) {
  l := CombinedLog
  r, err := http.NewRequest("GET", "http://golang.org", nil)
  if err != nil {
    t.Errorf("Failed to create request: %s", err)
  }
  r.RemoteAddr = "127.0.0.1"
  r.Header.Set("User-Agent", "Apache-LogFormat Port In Golang")
  r.Header.Set("Referer", "http://dummy.com")

  for i := 0; i < 100000; i ++ {
    output := l.FormatLoop(
      r,
      200,
      http.Header{ "Content-Type": []string{"text/plain"} },
      1500000,
    )
    if output == "" {
      t.Errorf("Failed to Format")
    }
  }
}

