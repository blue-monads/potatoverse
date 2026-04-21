package worderd

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/blue-monads/potatoverse/backend/services/datahub"
	"github.com/blue-monads/potatoverse/backend/services/signer"
	"github.com/blue-monads/potatoverse/backend/xtypes"
	"github.com/gin-gonic/gin"
)

type mockApp struct {
	xtypes.App
}

func (m *mockApp) Database() datahub.Database { return nil }
func (m *mockApp) Signer() *signer.Signer     { return nil }

func TestWorderdExecutor(t *testing.T) {
	if _, err := os.Stat(WorkerdBinary); os.IsNotExist(err) {
		t.Skip("workerd binary not found")
	}

	builder := &WorderdExecutorBuilder{app: &mockApp{}}

	code := `
export default {
  fetch(request, env, ctx) {
    return new Response("Hello from workerd!");
  }
};
`
	opt := &xtypes.ExecutorBuilderOption{
		SpaceId: 1,
		CodeLoader: func() (string, error) {
			return code, nil
		},
	}

	executor, err := builder.Build(opt)
	if err != nil {
		t.Fatalf("Failed to build executor: %v", err)
	}
	defer executor.Cleanup()

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		c, _ := gin.CreateTestContext(rw)
		c.Request = req
		event := &xtypes.HttpEvent{
			Request: c,
		}
		executor.HandleHttp(event)
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "Hello from workerd!" {
		t.Errorf("Expected 'Hello from workerd!', got '%s'", string(body))
	}

	fmt.Println("Test passed!")
}
