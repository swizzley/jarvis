package web

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	exception "github.com/blendlabs/go-exception"
	logger "github.com/blendlabs/go-logger"
)

const (
	// EnvironmentVariableBindAddr is an env var that determines (if set) what the bind address should be.
	EnvironmentVariableBindAddr = "BIND_ADDR"

	// EnvironmentVariablePort is an env var that determines what the default bind address port segment returns.
	EnvironmentVariablePort = "PORT"

	// EnvironmentVariableTLSCert is an env var that contains the TLS cert.
	EnvironmentVariableTLSCert = "TLS_CERT"

	// EnvironmentVariableTLSKey is an env var that contains the TLS key.
	EnvironmentVariableTLSKey = "TLS_KEY"

	// EnvironmentVariableTLSCertFile is an env var that contains the file path to the TLS cert.
	EnvironmentVariableTLSCertFile = "TLS_CERT_FILE"

	// EnvironmentVariableTLSKeyFile is an env var that contains the file path to the TLS key.
	EnvironmentVariableTLSKeyFile = "TLS_KEY_FILE"

	// DefaultPort is the default port the server binds to.
	DefaultPort = "8080"
)

// New returns a new app.
func New() *App {
	return &App{
		staticRewriteRules:    map[string][]*RewriteRule{},
		staticHeaders:         map[string]http.Header{},
		tlsCertLock:           &sync.Mutex{},
		auth:                  NewAuthManager(),
		viewCache:             NewViewCache(),
		readTimeout:           5 * time.Second,
		redirectTrailingSlash: true,
	}
}

// AppStartDelegate is a function that is run on start. Typically you use this to initialize the app.
type AppStartDelegate func(app *App) error

// App is the server for the app.
type App struct {
	name     string
	domain   string
	bindAddr string
	port     string

	logger *logger.Agent
	config interface{}

	tlsCertBytes, tlsKeyBytes []byte
	tlsCertLock               *sync.Mutex
	tlsCert                   *tls.Certificate

	startDelegate AppStartDelegate

	staticRewriteRules map[string][]*RewriteRule
	staticHeaders      map[string]http.Header

	routes                  map[string]*node
	notFoundHandler         Handler
	methodNotAllowedHandler Handler
	panicHandler            PanicHandler
	panicAction             PanicAction
	redirectTrailingSlash   bool
	handleOptions           bool
	handleMethodNotAllowed  bool

	defaultMiddleware []Middleware

	viewCache *ViewCache

	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration

	auth *AuthManager

	tx *sql.Tx
}

// Name returns the app name.
func (a *App) Name() string {
	return a.name
}

// SetName sets a log label for the app.
func (a *App) SetName(name string) {
	a.name = name
	if a.logger != nil {
		a.logger.Writer().SetLabel(name)
	}
}

// Domain returns the domain for the app.
func (a *App) Domain() string {
	return a.domain
}

// SetDomain sets the domain for the app.
func (a *App) SetDomain(domain string) {
	a.domain = domain
}

// ReadTimeout returns the read timeout for the server.
func (a *App) ReadTimeout() time.Duration {
	return a.readTimeout
}

// SetReadTimeout sets the read timeout for the server.
func (a *App) SetReadTimeout(readTimeout time.Duration) {
	a.readTimeout = readTimeout
}

// WriteTimeout returns the write timeout for the server.
func (a *App) WriteTimeout() time.Duration {
	return a.writeTimeout
}

// SetWriteTimeout sets the write timeout for the server.
func (a *App) SetWriteTimeout(writeTimeout time.Duration) {
	a.writeTimeout = writeTimeout
}

// UseTLS sets the app to use TLS.
func (a *App) UseTLS(tlsCert, tlsKey []byte) {
	a.tlsCertBytes = tlsCert
	a.tlsKeyBytes = tlsKey

	// this defaults to inferred or true.
	a.auth.SetCookieAsSecure(true)
}

// UseTLSFromEnvironment reads TLS settings from the environment.
func (a *App) UseTLSFromEnvironment() error {
	tlsCert := os.Getenv(EnvironmentVariableTLSCert)
	tlsKey := os.Getenv(EnvironmentVariableTLSKey)
	tlsCertPath := os.Getenv(EnvironmentVariableTLSCertFile)
	tlsKeyPath := os.Getenv(EnvironmentVariableTLSKeyFile)

	if len(tlsCert) > 0 && len(tlsKey) > 0 {
		a.UseTLS([]byte(tlsCert), []byte(tlsKey))
	} else if len(tlsCertPath) > 0 && len(tlsKeyPath) > 0 {
		cert, err := ioutil.ReadFile(tlsCertPath)
		if err != nil {
			return exception.Wrap(err)
		}

		key, err := ioutil.ReadFile(tlsKeyPath)
		if err != nil {
			return exception.Wrap(err)
		}

		a.UseTLS(cert, key)
	}
	return nil
}

// Logger returns the diagnostics agent for the app.
func (a *App) Logger() *logger.Agent {
	return a.logger
}

// SetLogger sets the diagnostics agent.
func (a *App) SetLogger(agent *logger.Agent) {
	a.logger = agent
	if a.logger != nil {
		a.logger.AddEventListener(logger.EventWebRequestStart, a.onRequestStart)
		a.logger.AddEventListener(logger.EventWebRequestPostBody, a.onRequestPostBody)
		a.logger.AddEventListener(logger.EventWebRequest, a.onRequestComplete)
		a.logger.AddEventListener(logger.EventWebResponse, a.onResponse)
	}
}

// Config returns the app config object.
func (a *App) Config() interface{} {
	return a.config
}

// SetConfig sets the app config object.
func (a *App) SetConfig(config interface{}) {
	a.config = config
}

// Auth returns the session manager.
func (a *App) Auth() *AuthManager {
	return a.auth
}

// SetAuth sets the session manager.
func (a *App) SetAuth(auth *AuthManager) {
	a.auth = auth
}

// InitializeConfig reads a config prototype from the environment.
func (a *App) InitializeConfig(configPrototype interface{}) error {
	config, err := ReadConfigFromEnvironment(configPrototype)
	if err != nil {
		return err
	}
	a.config = config
	return nil
}

func (a *App) onRequestStart(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
	if len(state) < 1 {
		return
	}
	context, isContext := state[0].(*Ctx)
	if !isContext {
		return
	}
	logger.WriteRequestStart(writer, ts, context.Request)
}

func (a *App) onRequestPostBody(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
	if len(state) < 1 {
		return
	}

	body, isBody := state[0].([]byte)
	if !isBody {
		return
	}

	logger.WriteRequestBody(writer, ts, body)
}

func (a *App) onRequestComplete(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
	if len(state) < 1 {
		return
	}
	context, isContext := state[0].(*Ctx)
	if !isContext {
		return
	}
	logger.WriteRequest(writer, ts, context.Request, context.Response.StatusCode(), context.Response.ContentLength(), context.Elapsed())
}

func (a *App) onResponse(writer logger.Logger, ts logger.TimeSource, eventFlag logger.EventFlag, state ...interface{}) {
	if len(state) < 1 {
		return
	}
	body, stateIsBody := state[0].([]byte)
	if !stateIsBody {
		return
	}
	logger.WriteResponseBody(writer, ts, body)
}

// IsolateTo sets the app to use a transaction for *all* requests.
// Caveat: only use during testing.
func (a *App) IsolateTo(tx *sql.Tx) {
	a.tx = tx
}

// Tx returns the isolated transaction.
func (a *App) Tx() *sql.Tx {
	return a.tx
}

// SetPort sets the port the app listens on.
// If BindAddr is not set, this will be returned in the form
// :Port(), as a result the server will bind to all available interfaces.
func (a *App) SetPort(port string) {
	a.port = port
}

// Port returns the port for the app.
// Port is the last in precedence behind BindAddr() for what
// ultimately forms the bind address the server binds to.
func (a *App) Port() string {
	if len(a.port) != 0 {
		return a.port
	}
	envVar := os.Getenv(EnvironmentVariablePort)
	if len(envVar) != 0 {
		return envVar
	}

	return DefaultPort
}

// SetDefaultMiddleware sets the application wide default middleware.
func (a *App) SetDefaultMiddleware(middleware ...Middleware) {
	a.defaultMiddleware = middleware
}

// DefaultMiddleware returns the default middleware.
func (a *App) DefaultMiddleware() []Middleware {
	return a.defaultMiddleware
}

// OnStart lets you register a task that is run before the server starts.
// Typically this delegate sets up the database connection and other init items.
func (a *App) OnStart(action AppStartDelegate) {
	a.startDelegate = action
}

// SetBindAddr sets the bind address of the server.
// It is the first in order of precedence for what ultimately will
// form the bind address that the server binds to.
func (a *App) SetBindAddr(bindAddr string) {
	a.bindAddr = bindAddr
}

// BindAddr returns the address the server will bind to.
func (a *App) BindAddr() string {
	if len(a.bindAddr) > 0 {
		return a.bindAddr
	}

	envVar := os.Getenv(EnvironmentVariableBindAddr)
	if len(envVar) > 0 {
		return envVar
	}

	return fmt.Sprintf(":%s", a.Port())
}

// Server returns the basic http.Server for the app.
func (a *App) Server() *http.Server {

	return &http.Server{
		Addr:              a.BindAddr(),
		Handler:           a,
		ReadTimeout:       a.readTimeout,
		ReadHeaderTimeout: a.readHeaderTimeout,
		WriteTimeout:      a.writeTimeout,
		IdleTimeout:       a.idleTimeout,
	}
}

// Start starts the server and binds to the given address.
func (a *App) Start() error {
	return a.StartWithServer(a.Server())
}

func (a *App) getCertificate(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if a.tlsCert == nil {
		tlsCert, err := tls.X509KeyPair(a.tlsCertBytes, a.tlsKeyBytes)
		if err != nil {
			return nil, err
		}
		a.tlsCert = &tlsCert
	}
	return a.tlsCert, nil
}

func (a *App) commonStartupTasks() error {
	return a.viewCache.Initialize()
}

// StartWithServer starts the app on a custom server.
// This lets you configure things like TLS keys and
// other options.
func (a *App) StartWithServer(server *http.Server) error {
	a.logger.OnEvent(EventAppStart, a)
	defer a.logger.OnEvent(EventAppExit, a)

	var err error
	if a.startDelegate != nil {
		a.logger.Infof("startup tasks starting")
		err = a.startDelegate(a)
		if err != nil {
			a.logger.Fatalf("startup tasks error: %v", err)
			return err
		}
		a.logger.Infof("startup tasks complete")
	}

	err = a.commonStartupTasks()
	if err != nil {
		a.logger.Fatalf("common startup tasks error: %v", err)
		return err
	}

	if a.logger != nil && a.logger.IsEnabled(logger.EventInfo) {
		serverProtocol := "http"
		if len(a.tlsCertBytes) > 0 && len(a.tlsKeyBytes) > 0 {
			serverProtocol = "https (tls)"
		}
		a.logger.Infof("%s server started, listening on %s", serverProtocol, server.Addr)
		a.logger.Infof("%s server diagnostics verbosity %s", serverProtocol, a.logger.Events().String())
	}

	a.logger.OnEvent(EventAppStartComplete, a)

	if len(a.tlsCertBytes) > 0 && len(a.tlsKeyBytes) > 0 {
		server.TLSConfig = &tls.Config{
			GetCertificate: a.getCertificate,
		}
		return exception.Wrap(server.ListenAndServeTLS("", ""))
	}

	return exception.Wrap(server.ListenAndServe())
}

// Register registers a controller with the app's router.
func (a *App) Register(c Controller) {
	c.Register(a)
}

func (a *App) middlewarePipeline(action Action, middleware ...Middleware) Action {
	finalMiddleware := make([]Middleware, len(middleware)+len(a.defaultMiddleware))
	cursor := len(finalMiddleware) - 1
	for i := len(a.defaultMiddleware) - 1; i >= 0; i-- {
		finalMiddleware[cursor] = a.defaultMiddleware[i]
		cursor--
	}

	for i := len(middleware) - 1; i >= 0; i-- {
		finalMiddleware[cursor] = middleware[i]
		cursor--
	}

	return NestMiddleware(action, finalMiddleware...)
}

// GET registers a GET request handler.
func (a *App) GET(path string, action Action, middleware ...Middleware) {
	a.handle("GET", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

// OPTIONS registers a OPTIONS request handler.
func (a *App) OPTIONS(path string, action Action, middleware ...Middleware) {
	a.handle("OPTIONS", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

// HEAD registers a HEAD request handler.
func (a *App) HEAD(path string, action Action, middleware ...Middleware) {
	a.handle("HEAD", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

// PUT registers a PUT request handler.
func (a *App) PUT(path string, action Action, middleware ...Middleware) {
	a.handle("PUT", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

// PATCH registers a PATCH request handler.
func (a *App) PATCH(path string, action Action, middleware ...Middleware) {
	a.handle("PATCH", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

// POST registers a POST request actions.
func (a *App) POST(path string, action Action, middleware ...Middleware) {
	a.handle("POST", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

// DELETE registers a DELETE request handler.
func (a *App) DELETE(path string, action Action, middleware ...Middleware) {
	a.handle("DELETE", path, a.renderAction(a.middlewarePipeline(action, middleware...)))
}

func (a *App) handle(method, path string, handler Handler) {
	if len(path) == 0 {
		panic("path must not be empty")
	}
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}
	if a.routes == nil {
		a.routes = make(map[string]*node)
	}

	root := a.routes[method]
	if root == nil {
		root = new(node)
		a.routes[method] = root
	}

	root.addRoute(method, path, handler)
}

func (a *App) allowed(path, reqMethod string) (allow string) {
	if path == "*" { // server-wide
		for method := range a.routes {
			if method == "OPTIONS" {
				continue
			}

			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
		return
	}
	for method := range a.routes {
		// Skip the requested method - we already tried this one
		if method == reqMethod || method == "OPTIONS" {
			continue
		}

		handle, _, _ := a.routes[method].getValue(path)
		if handle != nil {
			// add request method to list of allowed methods
			if len(allow) == 0 {
				allow = method
			} else {
				allow += ", " + method
			}
		}
	}
	if len(allow) > 0 {
		allow += ", OPTIONS"
	}
	return
}

// Lookup finds the route data for a given method and path.
func (a *App) Lookup(method, path string) (route *Route, params RouteParameters, slashRedirect bool) {
	if root := a.routes[method]; root != nil {
		return root.getValue(path)
	}
	return nil, nil, false
}

func (a *App) recv(w http.ResponseWriter, req *http.Request) {
	if rcv := recover(); rcv != nil {
		a.panicHandler(w, req, rcv)
	}
}

// ServeHTTP makes the router implement the http.Handler interface.
func (a *App) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if a.panicHandler != nil {
		defer a.recv(w, req)
	}

	path := req.URL.Path

	if root := a.routes[req.Method]; root != nil {
		if route, params, tsr := root.getValue(path); route != nil {
			route.Handler(w, req, route, params)
			return
		} else if req.Method != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if req.Method != "GET" {
				code = 307
			}

			if tsr && a.redirectTrailingSlash {
				if len(path) > 1 && path[len(path)-1] == '/' {
					req.URL.Path = path[:len(path)-1]
				} else {
					req.URL.Path = path + "/"
				}
				http.Redirect(w, req, req.URL.String(), code)
				return
			}
		}
	}

	if req.Method == "OPTIONS" {
		// Handle OPTIONS requests
		if a.handleOptions {
			if allow := a.allowed(path, req.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				return
			}
		}
	} else {
		// Handle 405
		if a.handleMethodNotAllowed {
			if allow := a.allowed(path, req.Method); len(allow) > 0 {
				w.Header().Set("Allow", allow)
				if a.methodNotAllowedHandler != nil {
					a.methodNotAllowedHandler(w, req, nil, nil)
				} else {
					http.Error(w,
						http.StatusText(http.StatusMethodNotAllowed),
						http.StatusMethodNotAllowed,
					)
				}
				return
			}
		}
	}

	// Handle 404
	if a.notFoundHandler != nil {
		a.notFoundHandler(w, req, nil, nil)
	} else {
		http.NotFound(w, req)
	}
}

// --------------------------------------------------------------------------------
// Static Result Methods
// --------------------------------------------------------------------------------

// AddStaticRewriteRule adds a rewrite rule for a specific statically served path.
// Make sure to serve the static path with app.Static(path, root).
func (a *App) AddStaticRewriteRule(path, match string, action RewriteAction) error {
	expr, err := regexp.Compile(match)
	if err != nil {
		return err
	}
	a.staticRewriteRules[path] = append(a.staticRewriteRules[path], &RewriteRule{
		MatchExpression: match,
		expr:            expr,
		Action:          action,
	})

	return nil
}

// AddStaticHeader adds a header for the given static path.
func (a *App) AddStaticHeader(path, key, value string) {
	if _, hasHeaders := a.staticHeaders[path]; !hasHeaders {
		a.staticHeaders[path] = http.Header{}
	}
	a.staticHeaders[path].Add(key, value)
}

// Static serves files from the given file system root.
// The path must end with "/*filepath", files are then served from the local
// path /defined/root/dir/*filepath.
// For example if root is "/etc" and *filepath is "passwd", the local file
// "/etc/passwd" would be served.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use http.Dir:
//     app.Static("/src/*filepath", http.Dir("/var/www"))
func (a *App) Static(path string, root http.FileSystem) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath in path '" + path + "'")
	}

	a.handle("GET", path, a.renderAction(a.staticAction(path, root)))
}

// staticAction returns a Action for a given static path and root.
func (a *App) staticAction(path string, root http.FileSystem) Action {
	fileServer := http.FileServer(root)

	return func(ctx *Ctx) Result {

		var staticRewriteRules []*RewriteRule
		var staticHeaders http.Header

		if rules, hasRules := a.staticRewriteRules[path]; hasRules {
			staticRewriteRules = rules
		}

		if headers, hasHeaders := a.staticHeaders[path]; hasHeaders {
			staticHeaders = headers
		}

		filePath, _ := ctx.RouteParam("filepath")

		return &StaticResult{
			FilePath:     filePath,
			FileServer:   fileServer,
			RewriteRules: staticRewriteRules,
			Headers:      staticHeaders,
		}
	}
}

// ViewCache returns the view result provider.
func (a *App) ViewCache() *ViewCache {
	return a.viewCache
}

// --------------------------------------------------------------------------------
// Router internal methods
// --------------------------------------------------------------------------------

// SetNotFoundHandler sets the not found handler.
func (a *App) SetNotFoundHandler(handler Action) {
	a.notFoundHandler = a.renderAction(handler)
}

// SetMethodNotAllowedHandler sets the not found handler.
func (a *App) SetMethodNotAllowedHandler(handler Action) {
	a.methodNotAllowedHandler = a.renderAction(handler)
}

// SetPanicHandler sets the not found handler.
func (a *App) SetPanicHandler(handler PanicAction) {
	a.panicAction = handler
	a.panicHandler = func(w http.ResponseWriter, r *http.Request, err interface{}) {
		a.renderAction(func(ctx *Ctx) Result {
			a.logger.ErrorEventWithState(logger.EventFatalError, logger.ColorRed, fmt.Errorf("%v", err), ctx)
			return handler(ctx, err)
		})(w, r, nil, nil)
	}
}

// --------------------------------------------------------------------------------
// Testing Methods
// --------------------------------------------------------------------------------

// Mock returns a request bulider to facilitate mocking requests.
func (a *App) Mock() *MockRequestBuilder {
	return NewMockRequestBuilder(a)
}

// --------------------------------------------------------------------------------
// Request Pipeline
// --------------------------------------------------------------------------------

// renderAction is the translation step from Action to Handler.
// this is where the bulk of the "pipeline" happens.
func (a *App) renderAction(action Action) Handler {
	return func(w http.ResponseWriter, r *http.Request, route *Route, p RouteParameters) {
		a.setResponseHeaders(w)
		response := a.newResponse(w, r)
		context := a.pipelineInit(response, r, route, p)
		a.renderResult(action, context)
		a.pipelineComplete(context)
	}
}

func (a *App) setResponseHeaders(w http.ResponseWriter) {
	w.Header().Set(HeaderServer, PackageName)
	w.Header().Set(HeaderXServedBy, PackageName)
}

func (a *App) newResponse(w http.ResponseWriter, r *http.Request) ResponseWriter {
	var response ResponseWriter
	if a.shouldCompressOutput(r) {
		w.Header().Set(HeaderContentEncoding, ContentEncodingGZIP)
		if a.logger.IsEnabled(logger.EventWebResponse) {
			response = NewBufferedCompressedResponseWriter(w)
		} else {
			response = NewCompressedResponseWriter(w)
		}
	} else {
		w.Header().Set(HeaderContentEncoding, ContentEncodingIdentity)
		if a.logger.IsEnabled(logger.EventWebResponse) {
			response = NewBufferedResponseWriter(w)
		} else {
			response = NewResponseWriter(w)
		}
	}
	return response
}

func (a *App) shouldCompressOutput(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}

func (a *App) pipelineInit(w ResponseWriter, r *http.Request, route *Route, p RouteParameters) *Ctx {
	context := a.newCtx(w, r, route, p)
	context.onRequestStart()
	a.logger.OnEvent(logger.EventWebRequestStart, context)
	return context
}

// Ctx creates a context.
func (a *App) newCtx(w ResponseWriter, r *http.Request, route *Route, p RouteParameters) *Ctx {
	ctx := NewCtx(w, r, p)
	ctx.app = a
	ctx.route = route
	ctx.auth = a.auth
	ctx.tx = a.tx
	ctx.logger = a.logger
	ctx.config = a.config

	// it is assumed that default middleware will override this at some point.
	ctx.SetDefaultResultProvider(ctx.Text())

	return ctx
}

func (a *App) renderResult(action Action, ctx *Ctx) error {
	result := action(ctx)
	if result != nil {
		err := result.Render(ctx)
		if err != nil {
			a.logger.Error(err)
			return err
		}
	}
	return nil
}

func (a *App) pipelineComplete(ctx *Ctx) {
	err := ctx.Response.Flush()
	if err != nil && err != http.ErrBodyNotAllowed {
		a.logger.Error(err)
	}
	ctx.onRequestEnd()
	ctx.setLoggedStatusCode(ctx.Response.StatusCode())
	ctx.setLoggedContentLength(ctx.Response.ContentLength())
	if a.logger.IsEnabled(logger.EventWebResponse) {
		a.logger.OnEvent(logger.EventWebResponse, ctx.Response.Bytes())
	}

	// effectively "request complete"
	a.logger.OnEvent(logger.EventWebRequest, ctx)
	err = ctx.Response.Close()
	if err != nil && err != http.ErrBodyNotAllowed {
		a.logger.Error(err)
	}
}
