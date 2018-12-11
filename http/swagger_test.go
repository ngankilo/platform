package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/influxdata/platform"
	"github.com/influxdata/platform/inmem"
	"github.com/influxdata/platform/kit/prom"
	"github.com/influxdata/platform/mock"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap/zaptest"
)

var (
	id         = platform.ID(100)
	orgName    = "org1"
	bucketName = "bucket1"
	userName   = "user1"
	password   = "hunter2"
	token      = "secret"
)

const (
	// PreconditionExt is used to search for operations
	// that must run as preconditons for an operation.
	PreconditionExt = "x-precondition-operations"
)

// TODO: recurse
func findOperations(ids []string, pathItems []pathItem) []pathItem {
	res := make([]pathItem, len(ids))
	for _, p := range pathItems {
		if p.Op.OperationID == "" {
			continue
		}

		for i, id := range ids {
			if p.Op.OperationID == id {
				res[i] = p
			}
		}
	}

	return res
}

func findPreconditions(p pathItem, pathItems []pathItem) []pathItem {
	o, ok := p.Op.Extensions[PreconditionExt]
	if !ok {
		return nil
	}
	ops := []string{}
	err := json.Unmarshal(o.(json.RawMessage), &ops)
	if err != nil {
		return nil
	}

	return findOperations(ops, pathItems)
}

type pathItem struct {
	Path string
	Verb string
	Op   *openapi3.Operation
}

func (p pathItem) request(t *testing.T, url string) (*http.Request, error) {
	t.Helper()
	path, err := p.requestPath(t, url)
	if err != nil {
		return nil, err
	}

	body, err := p.requestBody(t)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(p.Verb, path, body)
}

func (p pathItem) requestPath(t *testing.T, basePath string) (path string, err error) {
	path = basePath + p.Path
	for _, param := range p.Op.Parameters {
		switch param.Value.In {
		case openapi3.ParameterInPath:
			path = p.replacePathParam(t, path, param.Value)
		case openapi3.ParameterInQuery:
			path, err = p.addQueryParam(t, path, param.Value)
			if err != nil {
				return
			}
		}
	}
	u, err := url.Parse(path)
	if err != nil {
		return
	}
	return u.String(), err
}

func (p pathItem) replacePathParam(t *testing.T, path string, param *openapi3.Parameter) string {
	if param.Example == nil {
		return path
	}
	example, ok := param.Example.(string)
	if !ok {
		t.Errorf("invalid path parameter type: %t", param.Example)
		return path
	}
	name := "{" + param.Name + "}"
	path = strings.Replace(path, name, example, 1)
	return path
}

func (p pathItem) addQueryParam(t *testing.T, path string, param *openapi3.Parameter) (string, error) {
	if param.Example == nil {
		return path, nil
	}

	u, err := url.Parse(path)
	if err != nil {
		return path, err
	}
	query := u.Query()
	example := fmt.Sprintf("%v", param.Example)
	query.Add(param.Name, example)
	u.RawQuery = query.Encode()
	return u.String(), nil
}

func (p pathItem) requestBody(t *testing.T) (io.Reader, error) {
	var body io.Reader
	if p.Op.RequestBody != nil &&
		p.Op.RequestBody.Value != nil &&
		p.Op.RequestBody.Value.Content != nil &&
		p.Op.RequestBody.Value.Required {
		if len(p.Op.RequestBody.Value.Content) > 1 {
			t.Logf("multiple request value conent types")
		}
		for _, content := range p.Op.RequestBody.Value.Content {
			if content.Example == nil {
				continue
			}
			// TODO(goller): might be more interesting if we had named examples.
			b, err := json.Marshal(content.Example)
			if err != nil {
				t.Errorf("unable to marshal example response: %v", err)
				return nil, err
			}
			body = bytes.NewBuffer(b)
		}
	}
	return body, nil
}

func (p pathItem) responseBody(t *testing.T, status int, contentType string) []byte {
	t.Helper()
	if p.Op.Responses == nil {
		return nil
	}

	res, ok := p.Op.Responses[strconv.Itoa(status)]
	if !ok {
		res = p.Op.Responses["default"]
	}

	if res == nil || res.Value == nil || res.Value.Content == nil {
		return nil
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Errorf("unable to parse contentType '%v': %v", contentType, err)
		return nil
	}

	content, ok := res.Value.Content[mediaType]
	if !ok {
		// TODO: uncomment
		//t.Errorf("unknown mediaType %s", mediaType)
		return nil
	}
	if content == nil {
		return nil
	}
	if content.Example == nil {
		return nil
	}

	// TODO(goller): might be more interesting if we had named examples.
	b, err := json.Marshal(content.Example)
	if err != nil {
		t.Errorf("unable to marshal example response: %v", err)
		return nil
	}

	return b
}

func TestValidSwagger(t *testing.T) {
	data, err := ioutil.ReadFile("./swagger.yml")
	if err != nil {
		t.Fatalf("unable to read swagger specification: %v", err)
	}
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromYAMLData(data)
	if err != nil {
		t.Fatalf("unable to load swagger specification: %v", err)
	}
	if err := swagger.Validate(context.Background()); err != nil {
		t.Errorf("invalid swagger specification: %v", err)
	}
}

// LOGIC:
// Perhaps the swagger test will look up each
// route of the swagger doc in an initialization function.
// Then, initialization is run
// Then test is run against the entire server.
// then fill in the parameters using examples
// that'll setup the basis of an HTTP request
// fill in the request body if it exists
// make request
// get response and compare response to example response related to status
// Perhaps allow a custom response checker (useful for date/times)

// types of tests
// 1. if it is a resource then it should hae the standard CRUD operations
// 2. Everything has a description
// 3. Check consistency of status messages
// 4. everything should have a 404 if not found
// 5. every endpoint in the router has a route in swagger
// 6. resource endpoints should be plural named

func doRequest(t *testing.T, ts *httptest.Server, token string, item pathItem) (*http.Response, error) {
	req, err := item.request(t, ts.URL)
	if err != nil {
		t.Errorf("unable to create request: %v", err)
		return nil, err
	}
	validate := &openapi3filter.RequestValidationInput{
		Request: req,
		//PathParams: pathParams,
		//Route:      route,
	}
	if err := openapi3filter.ValidateRequest(context.Background(), validate); err != nil {
		// TODO: add validation
		//t.Errorf("request does not satisfy json schema: %v", err)
		//return nil, err
	}
	SetToken(token, req)
	return ts.Client().Do(req)
}

func TestHowdy(t *testing.T) {
	swagger := initSwagger(t)
	pathItems := swaggerPathItems(t, swagger)

	for _, item := range pathItems {
		t.Run(item.Path+" "+item.Verb, func(t *testing.T) {
			if strings.HasPrefix(item.Path, "/api/v2/macros") && item.Verb == "DELETE" {
				t.Skipf("skipping test %s as macro service has panics", t.Name())
				return
			}
			if item.Path == "/api/v2/sources" && item.Verb == "GET" {
				t.Skipf("skipping test %s as sources service has panics", t.Name())
				return
			}
			if item.Path == "/api/v2/tasks" && item.Verb == "GET" {
				t.Skipf("skipping test %s as tasks service has panics", t.Name())
				return
			}
			ts := initServer(t)
			defer ts.Close()

			conds := findPreconditions(item, pathItems)
			for _, cond := range conds {
				res, err := doRequest(t, ts, token, cond)
				if err != nil {
					t.Errorf("failed to request precondition: %v", err)
					return
				}
				if res.StatusCode >= 400 {
					t.Errorf("precondition failed: %s %s", cond.Path, cond.Verb)
					for _, cond := range conds {
						t.Errorf("preconditions: %s %s", cond.Path, cond.Verb)
					}
					t.Logf("Status: %d", res.StatusCode)
					t.Logf("Headers %v", res.Header)
					return
				}
			}
			res, err := doRequest(t, ts, token, item)
			if err != nil {
				t.Errorf("unable to retrieve http response: %v", err)
				return
			}
			body, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				t.Logf("Status: %d", res.StatusCode)
				t.Logf("Headers %v", res.Header)
				t.Errorf("unable to read body: %v", err)
				return
			}

			mediaType := res.Header.Get("Content-Type")
			example := item.responseBody(t, res.StatusCode, mediaType)
			if len(example) == 0 && len(body) == 0 {
				t.Logf("Status: %d", res.StatusCode)
				t.Logf("Headers %v", res.Header)
				t.Logf("no example and no body: %s", t.Name())
				return
			}
			got := strings.TrimSpace(string(body))
			want := string(example)
			if eq, _ := jsonEqual(got, want); !eq {
				t.Logf("Status: %d", res.StatusCode)
				t.Logf("Headers %v", res.Header)
				// TODO: uncomment...
				t.Errorf("\n***%s***\n,\nwant\n***%s***", got, want)
			}
		})
	}
}

func initServer(t *testing.T) *httptest.Server {
	t.Helper()
	logger := zaptest.NewLogger(t)
	reg := prom.NewRegistry()
	reg.MustRegister(prometheus.NewGoCollector())
	reg.WithLogger(logger)

	svc := inmem.NewService()
	svc.IDGenerator = &mock.IDGenerator{
		IDFn: func() platform.ID {
			return id
		},
	}
	svc.TokenGenerator = mock.NewTokenGenerator("secret", nil)

	ctx := context.Background()
	svc.PutUser(ctx, &platform.User{
		ID:   id,
		Name: userName,
	})
	svc.SetPassword(ctx, userName, password)
	svc.PutOrganization(ctx, &platform.Organization{
		ID:   id,
		Name: orgName,
	})
	svc.PutBucket(ctx, &platform.Bucket{
		ID:              id,
		Name:            bucketName,
		Organization:    orgName,
		OrganizationID:  id,
		RetentionPeriod: time.Hour,
	})

	svc.PutAuthorization(ctx, &platform.Authorization{
		ID:     platform.ID(1),
		Token:  token,
		Status: platform.Active,
		User:   userName,
		UserID: id,
		Permissions: []platform.Permission{
			platform.CreateUserPermission,
			platform.DeleteUserPermission,
			platform.Permission{
				Resource: platform.OrganizationResource,
				Action:   platform.WriteAction,
			},
			platform.WriteBucketPermission(id),
		},
	})
	svc.PutOnboardingStatus(ctx, true)

	handlerConfig := &APIBackend{
		Logger: logger,
		/*
		   NewBucketService:                source.NewBucketService,
		   NewQueryService:                 source.NewQueryService,
		   PointsWriter:                    pointsWriter,
		*/
		AuthorizationService: svc,
		BucketService:        svc,
		//SessionService:                  svc,
		UserService:                svc,
		OrganizationService:        svc,
		UserResourceMappingService: svc,
		DashboardService:           svc,
		//DashboardOperationLogService:    svc,
		//BucketOperationLogService:       svc,
		//UserOperationLogService:         svc,
		//OrganizationOperationLogService: svc,
		ViewService: svc,
		//SourceService:                   svc,
		MacroService:              svc,
		BasicAuthService:          svc,
		OnboardingService:         svc,
		ScraperTargetStoreService: svc,
		TelegrafService:           svc,
		//ProxyQueryService:               storageQueryService,
		//TaskService:                     taskSvc,
		//ChronografService:               chronografSvc,
	}

	h := NewHandlerFromRegistry("platform", reg)
	p := NewPlatformHandler(handlerConfig)
	reg.MustRegister(p.PrometheusCollectors()...)
	h.Handler = p
	ts := httptest.NewServer(h)
	return ts
}

func initSwagger(t *testing.T) *openapi3.Swagger {
	t.Helper()
	data, err := ioutil.ReadFile("./swagger.yml")
	if err != nil {
		t.Fatalf("unable to read swagger specification: %v", err)
	}
	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromYAMLData(data)
	if err != nil {
		t.Fatalf("unable to load swagger specification: %v", err)
	}

	//var router = openapi3filter.NewRouter().WithSwagger(swagger)
	return swagger
}

// swaggerPathItems returns
func swaggerPathItems(t *testing.T, swagger *openapi3.Swagger) []pathItem {
	t.Helper()

	bp := basePath("", &swagger.Servers)
	pathItems := []pathItem{}
	for p, item := range swagger.Paths {
		for verb, op := range item.Operations() {
			pathItems = append(pathItems, pathItem{
				Path: path.Clean(basePath(bp, op.Servers) + p),
				Verb: verb,
				Op:   op,
			})
		}
	}

	sort.Slice(pathItems, func(i, j int) bool {
		if pathItems[i].Path == pathItems[j].Path {
			return pathItems[i].Verb < pathItems[j].Verb
		}
		return pathItems[i].Path < pathItems[j].Path
	})
	return pathItems
}

func basePath(defaultPath string, servers *openapi3.Servers) string {
	if servers == nil {
		return defaultPath
	}
	for _, server := range *servers {
		if server.URL != "" {
			return server.URL
		}
	}
	return defaultPath
}
