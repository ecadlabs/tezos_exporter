package tezos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

const (
	libraryVersion = "0.0.1"
	userAgent      = "go-tezos/" + libraryVersion
	mediaType      = "application/json"
)

const (
	// ErrorKindPermanent Tezos RPC error kind.
	ErrorKindPermanent = "permanent"
	// ErrorKindTemporary Tezos RPC error kind.
	ErrorKindTemporary = "temporary"
	// ErrorKindBranch Tezos RPC error kind.
	ErrorKindBranch = "branch"
)

// HTTPError retains HTTP status
type HTTPError interface {
	error
	Status() string  // e.g. "200 OK"
	StatusCode() int // e.g. 200
	Body() []byte
}

// RPCError is a Tezos RPC error as documented on http://tezos.gitlab.io/mainnet/api/errors.html.
type RPCError interface {
	HTTPError
	ID() string
	Kind() string // e.g. "permanent"
	Raw() map[string]interface{}
	Errors() []RPCError // returns all errors as a slice
}

type httpError struct {
	status     string
	statusCode int
	body       []byte
}

func (e *httpError) Error() string {
	return fmt.Sprintf("tezos: HTTP status %v", e.statusCode)
}

func (e *httpError) Status() string {
	return e.status
}

func (e *httpError) StatusCode() int {
	return e.statusCode
}

func (e *httpError) Body() []byte {
	return e.body
}

type rpcError struct {
	*httpError
	id   string
	kind string // e.g. "permanent"
	raw  map[string]interface{}
}

func (e *rpcError) Error() string {
	return fmt.Sprintf("tezos: RPC error (kind = %q, id = %q)", e.kind, e.id)
}

func (e *rpcError) ID() string {
	return e.id
}

func (e *rpcError) Kind() string {
	return e.kind
}

func (e *rpcError) Raw() map[string]interface{} {
	return e.raw
}

func (e *rpcError) Errors() []RPCError {
	return []RPCError{e}
}

type rpcErrors struct {
	*httpError
	errors []*rpcError
}

func (e *rpcErrors) Error() string {
	if len(e.errors) == 0 {
		return ""
	}
	return e.errors[0].Error()
}

func (e *rpcErrors) ID() string {
	if len(e.errors) == 0 {
		return ""
	}
	return e.errors[0].id
}

func (e *rpcErrors) Kind() string {
	if len(e.errors) == 0 {
		return ""
	}
	return e.errors[0].kind
}

func (e *rpcErrors) Raw() map[string]interface{} {
	if len(e.errors) == 0 {
		return nil
	}
	return e.errors[0].raw
}

func (e *rpcErrors) Errors() []RPCError {
	res := make([]RPCError, len(e.errors))
	for i := range e.errors {
		res[i] = e.errors[i]
	}
	return res
}

type plainError struct {
	*httpError
	msg string
}

func (e *plainError) Error() string {
	return e.msg
}

var (
	_ RPCError = &rpcErrors{}
	_ RPCError = &rpcError{}
)

// NewRequest creates a Tezos RPC request.
func (c *RPCClient) NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	buf := new(bytes.Buffer)
	if body != nil {
		err = json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", mediaType)
	req.Header.Add("Accept", mediaType)
	req.Header.Add("User-Agent", c.UserAgent)

	if ctx != nil {
		return req.WithContext(ctx), nil
	}

	return req, nil
}

// RPCClient manages communication with a Tezos RPC server.
type RPCClient struct {
	// HTTP client used to communicate with the Tezos node API.
	client *http.Client
	// Base URL for API requests.
	BaseURL *url.URL
	// User agent name for client.
	UserAgent string
}

// NewRPCClient returns a new Tezos RPC client.
func NewRPCClient(httpClient *http.Client, baseURL string) (*RPCClient, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	c := &RPCClient{client: httpClient, BaseURL: u, UserAgent: userAgent}
	return c, nil
}

func (c *RPCClient) handleNormalResponse(ctx context.Context, resp *http.Response, v interface{}) error {
	// Normal return
	dec := json.NewDecoder(resp.Body)
	typ := reflect.TypeOf(v)

	if typ.Kind() == reflect.Chan {
		// Handle channel
		cases := []reflect.SelectCase{
			reflect.SelectCase{
				Dir:  reflect.SelectSend,
				Chan: reflect.ValueOf(v),
			},
			reflect.SelectCase{
				Dir:  reflect.SelectRecv,
				Chan: reflect.ValueOf(ctx.Done()),
			},
		}

		for {
			chunkVal := reflect.New(typ.Elem())

			if err := dec.Decode(chunkVal.Interface()); err != nil {
				if err == io.EOF {
					break
				}
				return err
			}

			cases[0].Send = chunkVal.Elem()
			if chosen, _, _ := reflect.Select(cases); chosen == 1 {
				return ctx.Err()
			}
		}

		return nil
	}

	// Handle single object
	if err := dec.Decode(&v); err != nil {
		return err
	}

	return nil
}

// Do retrieves values from the API and marshals them into the provided interface.
func (c *RPCClient) Do(req *http.Request, v interface{}) (err error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if rerr := resp.Body.Close(); err == nil {
			err = rerr
		}
	}()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}

	statusClass := resp.StatusCode / 100
	if statusClass == 2 {
		if v == nil {
			return nil
		}
		return c.handleNormalResponse(req.Context(), resp, v)
	}

	// Handle errors
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	httpErr := httpError{
		status:     resp.Status,
		statusCode: resp.StatusCode,
		body:       body,
	}

	if statusClass != 5 || !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		// Other errors with unknown body format (usually human readable string)
		return &httpErr
	}

	var raw interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return &plainError{&httpErr, fmt.Sprintf("tezos: error decoding RPC error: %v", err)}
	}

	// Can be an array
	var maps []map[string]interface{}

	switch e := raw.(type) {
	case []interface{}:
		for _, v := range e {
			if m, ok := v.(map[string]interface{}); ok {
				maps = append(maps, m)
			}
		}

	case map[string]interface{}:
		maps = []map[string]interface{}{e}

	default:
		return &plainError{&httpErr, "tezos: error decoding RPC error"}
	}

	if len(maps) == 0 {
		return &plainError{&httpErr, "tezos: empty error response"}
	}

	errs := make([]*rpcError, len(maps))

	for i, m := range maps {
		errID, ok := m["id"].(string)
		if !ok {
			return &plainError{&httpErr, "tezos: error decoding RPC error"}
		}

		errKind, ok := m["kind"].(string)
		if !ok {
			return &plainError{&httpErr, "tezos: error decoding RPC error"}
		}

		errs[i] = &rpcError{
			httpError: &httpErr,
			id:        errID,
			kind:      errKind,
			raw:       m,
		}
	}

	if len(errs) == 1 {
		return errs[0]
	}

	return &rpcErrors{
		httpError: &httpErr,
		errors:    errs,
	}
}
