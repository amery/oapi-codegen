// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/reedom/oapi-codegen version (devel) DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/reedom/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-chi/chi/v5"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Returns all pets
	// (GET /pets)
	FindPets(w http.ResponseWriter, r *http.Request, params FindPetsParams)
	// Creates a new pet
	// (POST /pets)
	AddPet(w http.ResponseWriter, r *http.Request)
	// Deletes a pet by ID
	// (DELETE /pets/{id})
	DeletePet(w http.ResponseWriter, r *http.Request, id int64)
	// Returns a pet by ID
	// (GET /pets/{id})
	FindPetByID(w http.ResponseWriter, r *http.Request, id int64)
}

// ServerInterfaceWrapper converts contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler            ServerInterface
	HandlerMiddlewares []MiddlewareFunc
	ErrorHandlerFunc   func(w http.ResponseWriter, r *http.Request, err error)
}

type MiddlewareFunc func(http.Handler) http.Handler

// FindPets operation middleware
func (siw *ServerInterfaceWrapper) FindPets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params FindPetsParams

	// ------------- Optional query parameter "tags" -------------

	err = runtime.BindQueryParameter("form", true, false, "tags", r.URL.Query(), &params.Tags)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "tags", Err: err})
		return
	}

	// ------------- Optional query parameter "limit" -------------

	err = runtime.BindQueryParameter("form", true, false, "limit", r.URL.Query(), &params.Limit)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "limit", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.FindPets(w, r, params)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// AddPet operation middleware
func (siw *ServerInterfaceWrapper) AddPet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.AddPet(w, r)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// DeletePet operation middleware
func (siw *ServerInterfaceWrapper) DeletePet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "id" -------------
	var id int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "id", runtime.ParamLocationPath, chi.URLParam(r, "id"), &id)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "id", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.DeletePet(w, r, id)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

// FindPetByID operation middleware
func (siw *ServerInterfaceWrapper) FindPetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var err error

	// ------------- Path parameter "id" -------------
	var id int64

	err = runtime.BindStyledParameterWithLocation("simple", false, "id", runtime.ParamLocationPath, chi.URLParam(r, "id"), &id)
	if err != nil {
		siw.ErrorHandlerFunc(w, r, &InvalidParamFormatError{ParamName: "id", Err: err})
		return
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		siw.Handler.FindPetByID(w, r, id)
	})

	for _, middleware := range siw.HandlerMiddlewares {
		handler = middleware(handler)
	}

	handler.ServeHTTP(w, r.WithContext(ctx))
}

type UnescapedCookieParamError struct {
	ParamName string
	Err       error
}

func (e *UnescapedCookieParamError) Error() string {
	return fmt.Sprintf("error unescaping cookie parameter '%s'", e.ParamName)
}

func (e *UnescapedCookieParamError) Unwrap() error {
	return e.Err
}

type UnmarshalingParamError struct {
	ParamName string
	Err       error
}

func (e *UnmarshalingParamError) Error() string {
	return fmt.Sprintf("Error unmarshaling parameter %s as JSON: %s", e.ParamName, e.Err.Error())
}

func (e *UnmarshalingParamError) Unwrap() error {
	return e.Err
}

type RequiredParamError struct {
	ParamName string
}

func (e *RequiredParamError) Error() string {
	return fmt.Sprintf("Query argument %s is required, but not found", e.ParamName)
}

type RequiredHeaderError struct {
	ParamName string
	Err       error
}

func (e *RequiredHeaderError) Error() string {
	return fmt.Sprintf("Header parameter %s is required, but not found", e.ParamName)
}

func (e *RequiredHeaderError) Unwrap() error {
	return e.Err
}

type InvalidParamFormatError struct {
	ParamName string
	Err       error
}

func (e *InvalidParamFormatError) Error() string {
	return fmt.Sprintf("Invalid format for parameter %s: %s", e.ParamName, e.Err.Error())
}

func (e *InvalidParamFormatError) Unwrap() error {
	return e.Err
}

type TooManyValuesForParamError struct {
	ParamName string
	Count     int
}

func (e *TooManyValuesForParamError) Error() string {
	return fmt.Sprintf("Expected one value for %s, got %d", e.ParamName, e.Count)
}

// Handler creates http.Handler with routing matching OpenAPI spec.
func Handler(si ServerInterface) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{})
}

type ChiServerOptions struct {
	BaseURL          string
	BaseRouter       chi.Router
	Middlewares      []MiddlewareFunc
	ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

// HandlerFromMux creates http.Handler with routing matching OpenAPI spec based on the provided mux.
func HandlerFromMux(si ServerInterface, r chi.Router) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseRouter: r,
	})
}

func HandlerFromMuxWithBaseURL(si ServerInterface, r chi.Router, baseURL string) http.Handler {
	return HandlerWithOptions(si, ChiServerOptions{
		BaseURL:    baseURL,
		BaseRouter: r,
	})
}

// HandlerWithOptions creates http.Handler with additional options
func HandlerWithOptions(si ServerInterface, options ChiServerOptions) http.Handler {
	r := options.BaseRouter

	if r == nil {
		r = chi.NewRouter()
	}
	if options.ErrorHandlerFunc == nil {
		options.ErrorHandlerFunc = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	wrapper := ServerInterfaceWrapper{
		Handler:            si,
		HandlerMiddlewares: options.Middlewares,
		ErrorHandlerFunc:   options.ErrorHandlerFunc,
	}

	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/pets", wrapper.FindPets)
	})
	r.Group(func(r chi.Router) {
		r.Post(options.BaseURL+"/pets", wrapper.AddPet)
	})
	r.Group(func(r chi.Router) {
		r.Delete(options.BaseURL+"/pets/{id}", wrapper.DeletePet)
	})
	r.Group(func(r chi.Router) {
		r.Get(options.BaseURL+"/pets/{id}", wrapper.FindPetByID)
	})

	return r
}

type FindPetsRequestObject struct {
	Params FindPetsParams
}

type FindPetsResponseObject interface {
	VisitFindPetsResponse(w http.ResponseWriter) error
}

type FindPets200JSONResponse []Pet

func (response FindPets200JSONResponse) VisitFindPetsResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type FindPetsdefaultJSONResponse struct {
	Body       Error
	StatusCode int
}

func (response FindPetsdefaultJSONResponse) VisitFindPetsResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)

	return json.NewEncoder(w).Encode(response.Body)
}

type AddPetRequestObject struct {
	Body *AddPetJSONRequestBody
}

type AddPetResponseObject interface {
	VisitAddPetResponse(w http.ResponseWriter) error
}

type AddPet200JSONResponse Pet

func (response AddPet200JSONResponse) VisitAddPetResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type AddPetdefaultJSONResponse struct {
	Body       Error
	StatusCode int
}

func (response AddPetdefaultJSONResponse) VisitAddPetResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)

	return json.NewEncoder(w).Encode(response.Body)
}

type DeletePetRequestObject struct {
	Id int64 `json:"id"`
}

type DeletePetResponseObject interface {
	VisitDeletePetResponse(w http.ResponseWriter) error
}

type DeletePet204Response struct {
}

func (response DeletePet204Response) VisitDeletePetResponse(w http.ResponseWriter) error {
	w.WriteHeader(204)
	return nil
}

type DeletePetdefaultJSONResponse struct {
	Body       Error
	StatusCode int
}

func (response DeletePetdefaultJSONResponse) VisitDeletePetResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)

	return json.NewEncoder(w).Encode(response.Body)
}

type FindPetByIDRequestObject struct {
	Id int64 `json:"id"`
}

type FindPetByIDResponseObject interface {
	VisitFindPetByIDResponse(w http.ResponseWriter) error
}

type FindPetByID200JSONResponse Pet

func (response FindPetByID200JSONResponse) VisitFindPetByIDResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)

	return json.NewEncoder(w).Encode(response)
}

type FindPetByIDdefaultJSONResponse struct {
	Body       Error
	StatusCode int
}

func (response FindPetByIDdefaultJSONResponse) VisitFindPetByIDResponse(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)

	return json.NewEncoder(w).Encode(response.Body)
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {
	// Returns all pets
	// (GET /pets)
	FindPets(ctx context.Context, request FindPetsRequestObject) (FindPetsResponseObject, error)
	// Creates a new pet
	// (POST /pets)
	AddPet(ctx context.Context, request AddPetRequestObject) (AddPetResponseObject, error)
	// Deletes a pet by ID
	// (DELETE /pets/{id})
	DeletePet(ctx context.Context, request DeletePetRequestObject) (DeletePetResponseObject, error)
	// Returns a pet by ID
	// (GET /pets/{id})
	FindPetByID(ctx context.Context, request FindPetByIDRequestObject) (FindPetByIDResponseObject, error)
}

type StrictHandlerFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request, args interface{}) (interface{}, error)

type StrictMiddlewareFunc func(f StrictHandlerFunc, operationID string) StrictHandlerFunc

type StrictHTTPServerOptions struct {
	RequestErrorHandlerFunc  func(w http.ResponseWriter, r *http.Request, err error)
	ResponseErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)
}

func NewStrictHandler(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares, options: StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusBadRequest)
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		},
	}}
}

func NewStrictHandlerWithOptions(ssi StrictServerInterface, middlewares []StrictMiddlewareFunc, options StrictHTTPServerOptions) ServerInterface {
	return &strictHandler{ssi: ssi, middlewares: middlewares, options: options}
}

type strictHandler struct {
	ssi         StrictServerInterface
	middlewares []StrictMiddlewareFunc
	options     StrictHTTPServerOptions
}

// FindPets operation middleware
func (sh *strictHandler) FindPets(w http.ResponseWriter, r *http.Request, params FindPetsParams) {
	var request FindPetsRequestObject

	request.Params = params

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.FindPets(ctx, request.(FindPetsRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "FindPets")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(FindPetsResponseObject); ok {
		if err := validResponse.VisitFindPetsResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", response))
	}
}

// AddPet operation middleware
func (sh *strictHandler) AddPet(w http.ResponseWriter, r *http.Request) {
	var request AddPetRequestObject

	var body AddPetJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		sh.options.RequestErrorHandlerFunc(w, r, fmt.Errorf("can't decode JSON body: %w", err))
		return
	}
	request.Body = &body

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.AddPet(ctx, request.(AddPetRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "AddPet")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(AddPetResponseObject); ok {
		if err := validResponse.VisitAddPetResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", response))
	}
}

// DeletePet operation middleware
func (sh *strictHandler) DeletePet(w http.ResponseWriter, r *http.Request, id int64) {
	var request DeletePetRequestObject

	request.Id = id

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.DeletePet(ctx, request.(DeletePetRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "DeletePet")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(DeletePetResponseObject); ok {
		if err := validResponse.VisitDeletePetResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", response))
	}
}

// FindPetByID operation middleware
func (sh *strictHandler) FindPetByID(w http.ResponseWriter, r *http.Request, id int64) {
	var request FindPetByIDRequestObject

	request.Id = id

	handler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, request interface{}) (interface{}, error) {
		return sh.ssi.FindPetByID(ctx, request.(FindPetByIDRequestObject))
	}
	for _, middleware := range sh.middlewares {
		handler = middleware(handler, "FindPetByID")
	}

	response, err := handler(r.Context(), w, r, request)

	if err != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, err)
	} else if validResponse, ok := response.(FindPetByIDResponseObject); ok {
		if err := validResponse.VisitFindPetByIDResponse(w); err != nil {
			sh.options.ResponseErrorHandlerFunc(w, r, err)
		}
	} else if response != nil {
		sh.options.ResponseErrorHandlerFunc(w, r, fmt.Errorf("Unexpected response type: %T", response))
	}
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"3d3v/hYAAP//wO3O5VcSAAA=",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
