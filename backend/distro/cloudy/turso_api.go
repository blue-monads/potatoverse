package cloudy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const tursoAPIBase = "https://api.turso.tech"

// tursoAPI is a thin client over the Turso platform API, used to provision one
// database per tenant. Turso Cloud has no namespaces: every database is routed
// by its own hostname, so a tenant database must exist before syncing to it.
type tursoAPI struct {
	token  string
	org    string
	group  string
	prefix string
	client *http.Client
}

type tursoDatabase struct {
	Name     string `json:"Name"`
	Hostname string `json:"Hostname"`
	Group    string `json:"group"`
}

type tenantRemote struct {
	URL       string
	AuthToken string
	Namespace string
}

func newTursoAPI(config *Config) *tursoAPI {
	if config.TursoAPIToken == "" || config.TursoOrg == "" {
		return nil
	}

	group := config.TursoGroup
	if group == "" {
		group = "default"
	}
	prefix := config.TursoDBPrefix
	if prefix == "" {
		prefix = "cloudy-"
	}

	return &tursoAPI{
		token:  config.TursoAPIToken,
		org:    config.TursoOrg,
		group:  group,
		prefix: prefix,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (t *tursoAPI) dbName(tenant string) (string, error) {
	name := t.prefix + tenant
	if len(name) > 64 {
		return "", fmt.Errorf("turso database name %q exceeds 64 characters", name)
	}
	return name, nil
}

func (t *tursoAPI) do(ctx context.Context, method, path string, body, out any) (int, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, tursoAPIBase+path, reader)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+t.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := t.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, err
	}
	if res.StatusCode >= 400 {
		return res.StatusCode, fmt.Errorf("turso api %s %s: %s: %s", method, path, res.Status, strings.TrimSpace(string(raw)))
	}
	if out != nil && len(raw) > 0 {
		if err := json.Unmarshal(raw, out); err != nil {
			return res.StatusCode, err
		}
	}
	return res.StatusCode, nil
}

func (t *tursoAPI) getDatabase(ctx context.Context, name string) (*tursoDatabase, error) {
	var out struct {
		Database tursoDatabase `json:"database"`
	}
	status, err := t.do(ctx, http.MethodGet,
		fmt.Sprintf("/v1/organizations/%s/databases/%s", t.org, name), nil, &out)
	if status == http.StatusNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &out.Database, nil
}

func (t *tursoAPI) createDatabase(ctx context.Context, name string) (*tursoDatabase, error) {
	var out struct {
		Database tursoDatabase `json:"database"`
	}
	_, err := t.do(ctx, http.MethodPost,
		fmt.Sprintf("/v1/organizations/%s/databases", t.org),
		map[string]string{"name": name, "group": t.group}, &out)
	if err != nil {
		return nil, err
	}
	return &out.Database, nil
}

func (t *tursoAPI) createToken(ctx context.Context, name string) (string, error) {
	var out struct {
		JWT string `json:"jwt"`
	}
	query := url.Values{"authorization": {"full-access"}}
	_, err := t.do(ctx, http.MethodPost,
		fmt.Sprintf("/v1/organizations/%s/databases/%s/auth/tokens?%s", t.org, name, query.Encode()),
		nil, &out)
	if err != nil {
		return "", err
	}
	if out.JWT == "" {
		return "", fmt.Errorf("turso api returned an empty token for %s", name)
	}
	return out.JWT, nil
}

// ensureTenantDB creates the tenant database if it does not exist yet and
// returns the remote URL plus a database scoped auth token.
func (t *tursoAPI) ensureTenantDB(ctx context.Context, tenant string) (tenantRemote, error) {
	name, err := t.dbName(tenant)
	if err != nil {
		return tenantRemote{}, err
	}

	db, err := t.getDatabase(ctx, name)
	if err != nil {
		return tenantRemote{}, err
	}
	if db == nil {
		db, err = t.createDatabase(ctx, name)
		if err != nil {
			return tenantRemote{}, err
		}
	}
	if db.Hostname == "" {
		return tenantRemote{}, fmt.Errorf("turso database %s has no hostname", name)
	}

	token, err := t.createToken(ctx, name)
	if err != nil {
		return tenantRemote{}, err
	}

	return tenantRemote{
		URL:       "https://" + db.Hostname,
		AuthToken: token,
	}, nil
}
