package consumer

import (
	"log"
	"testing"

	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/utils"
)

// V2HTTPMockProvider is the entrypoint for V2 http consumer tests
// This object is not thread safe
type V2HTTPMockProvider struct {
	*httpMockProvider
}

// NewV2Pact configures a new V2 HTTP Mock Provider for consumer tests
func NewV2Pact(config MockHTTPProviderConfig) (*V2HTTPMockProvider, error) {
	provider := &V2HTTPMockProvider{
		httpMockProvider: &httpMockProvider{
			config:               config,
			specificationVersion: models.V2,
		},
	}
	err := provider.configure()

	if err != nil {
		return nil, err
	}

	return provider, err
}

// AddInteraction to the pact
func (p *V2HTTPMockProvider) AddInteraction() *UnconfiguredV2Interaction {
	log.Println("[DEBUG] pact add V2 interaction")
	interaction := p.httpMockProvider.mockserver.NewInteraction("")

	i := &UnconfiguredV2Interaction{
		interaction: &Interaction{
			specificationVersion: models.V2,
			interaction:          interaction,
		},
		provider: p,
	}

	return i
}

type UnconfiguredV2Interaction struct {
	interaction *Interaction
	provider    *V2HTTPMockProvider
}

// Given specifies a provider state, may be called multiple times. Optional.
func (i *UnconfiguredV2Interaction) Given(state string) *UnconfiguredV2Interaction {
	i.interaction.interaction.Given(state)

	return i
}

type V2InteractionWithRequest struct {
	interaction *Interaction
	provider    *V2HTTPMockProvider
}

type V2RequestBuilder func(*V2InteractionWithRequestBuilder)

type V2InteractionWithRequestBuilder struct {
	interaction *Interaction
	provider    *V2HTTPMockProvider
}

// UponReceiving specifies the name of the test case. This becomes the name of
// the consumer/provider pair in the Pact file. Mandatory.
func (i *UnconfiguredV2Interaction) UponReceiving(description string) *UnconfiguredV2Interaction {
	i.interaction.interaction.UponReceiving(description)

	return i
}

// WithRequest provides a builder for the expected request
func (i *UnconfiguredV2Interaction) WithCompleteRequest(request Request) *V2InteractionWithCompleteRequest {
	i.interaction.WithCompleteRequest(request)

	return &V2InteractionWithCompleteRequest{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

type V2InteractionWithCompleteRequest struct {
	interaction *Interaction
	provider    *V2HTTPMockProvider
}

// WithRequest provides a builder for the expected request
func (i *V2InteractionWithCompleteRequest) WithCompleteResponse(response Response) *V2InteractionWithResponse {
	i.interaction.WithCompleteResponse(response)

	return &V2InteractionWithResponse{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

// WithRequest provides a builder for the expected request
func (i *UnconfiguredV2Interaction) WithRequest(method Method, path string, builders ...V2RequestBuilder) *V2InteractionWithRequest {
	return i.WithRequestPathMatcher(method, matchers.String(path), builders...)
}

// WithRequestPathMatcher allows a matcher in the expected request path
func (i *UnconfiguredV2Interaction) WithRequestPathMatcher(method Method, path matchers.Matcher, builders ...V2RequestBuilder) *V2InteractionWithRequest {
	i.interaction.interaction.WithRequest(string(method), path)

	for _, builder := range builders {
		builder(&V2InteractionWithRequestBuilder{
			interaction: i.interaction,
			provider:    i.provider,
		})
	}

	return &V2InteractionWithRequest{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

// Query specifies any query string on the expect request
func (i *V2InteractionWithRequestBuilder) Query(key string, values ...matchers.Matcher) *V2InteractionWithRequestBuilder {
	i.interaction.interaction.WithQuery(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Header adds a header to the expected request
func (i *V2InteractionWithRequestBuilder) Header(key string, values ...matchers.Matcher) *V2InteractionWithRequestBuilder {
	i.interaction.interaction.WithRequestHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Headers sets the headers on the expected request
func (i *V2InteractionWithRequestBuilder) Headers(headers matchers.HeadersMatcher) *V2InteractionWithRequestBuilder {
	i.interaction.interaction.WithRequestHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

// JSONBody adds a JSON body to the expected request
func (i *V2InteractionWithRequestBuilder) JSONBody(body interface{}) *V2InteractionWithRequestBuilder {
	// TODO: Don't like panic, but not sure if there is a better builder experience?
	if err := validateMatchers(i.interaction.specificationVersion, body); err != nil {
		panic(err)
	}

	if s, ok := body.(string); ok {
		// Check if someone tried to add an object as a string representation
		// as per original allowed implementation, e.g.
		// { "foo": "bar", "baz": like("bat") }
		if utils.IsJSONFormattedObject(string(s)) {
			log.Println("[WARN] request body appears to be a JSON formatted object, " +
				"no matching will occur. Support for structured strings has been" +
				"deprecated as of 0.13.0. Please use the JSON() method instead")
		}
	}

	i.interaction.interaction.WithJSONRequestBody(body)

	return i
}

// BinaryBody adds a binary body to the expected request
func (i *V2InteractionWithRequestBuilder) BinaryBody(body []byte) *V2InteractionWithRequestBuilder {
	i.interaction.interaction.WithBinaryRequestBody(body)

	return i
}

// MultipartBody adds a multipart  body to the expected request
func (i *V2InteractionWithRequestBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V2InteractionWithRequestBuilder {
	i.interaction.interaction.WithRequestMultipartFile(contentType, filename, mimePartName)

	return i
}

// Body adds general body to the expected request
func (i *V2InteractionWithRequestBuilder) Body(contentType string, body []byte) *V2InteractionWithRequestBuilder {
	// Check if someone tried to add an object as a string representation
	// as per original allowed implementation, e.g.
	// { "foo": "bar", "baz": like("bat") }
	if utils.IsJSONFormattedObject(string(body)) {
		log.Println("[WARN] request body appears to be a JSON formatted object, " +
			"no matching will occur. Support for structured strings has been" +
			"deprecated as of 0.13.0. Please use the JSON() method instead")
	}

	i.interaction.interaction.WithRequestBody(contentType, body)

	return i
}

// BodyMatch uses struct tags to automatically determine matchers from the given struct
func (i *V2InteractionWithRequestBuilder) BodyMatch(body interface{}) *V2InteractionWithRequestBuilder {
	i.interaction.interaction.WithJSONRequestBody(matchers.MatchV2(body))

	return i
}

// WillRespondWith sets the expected status and provides a response builder
func (i *V2InteractionWithRequest) WillRespondWith(status int, builders ...V2ResponseBuilder) *V2InteractionWithResponse {
	i.interaction.interaction.WithStatus(status)

	for _, builder := range builders {

		builder(&V2InteractionWithResponseBuilder{
			interaction: i.interaction,
			provider:    i.provider,
		})
	}

	return &V2InteractionWithResponse{
		interaction: i.interaction,
		provider:    i.provider,
	}
}

type V2ResponseBuilder func(*V2InteractionWithResponseBuilder)

type V2InteractionWithResponseBuilder struct {
	interaction *Interaction
	provider    *V2HTTPMockProvider
}

type V2InteractionWithResponse struct {
	interaction *Interaction
	provider    *V2HTTPMockProvider
}

// Header adds a header to the expected response
func (i *V2InteractionWithResponseBuilder) Header(key string, values ...matchers.Matcher) *V2InteractionWithResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(keyValuesToMapStringArrayInterface(key, values...))

	return i
}

// Headers sets the headers on the expected response
func (i *V2InteractionWithResponseBuilder) Headers(headers matchers.HeadersMatcher) *V2InteractionWithResponseBuilder {
	i.interaction.interaction.WithResponseHeaders(headersMatcherToNativeHeaders(headers))

	return i
}

// JSONBody adds a JSON body to the expected response
func (i *V2InteractionWithResponseBuilder) JSONBody(body interface{}) *V2InteractionWithResponseBuilder {
	// TODO: Don't like panic, how to build a better builder here - nil return + log?
	if err := validateMatchers(i.interaction.specificationVersion, body); err != nil {
		panic(err)
	}

	if s, ok := body.(string); ok {
		// Check if someone tried to add an object as a string representation
		// as per original allowed implementation, e.g.
		// { "foo": "bar", "baz": like("bat") }
		if utils.IsJSONFormattedObject(string(s)) {
			log.Println("[WARN] response body appears to be a JSON formatted object, " +
				"no matching will occur. Support for structured strings has been" +
				"deprecated as of 0.13.0. Please use the JSON() method instead")
		}
	}
	i.interaction.interaction.WithJSONResponseBody(body)

	return i
}

// BinaryBody adds a binary body to the expected response
func (i *V2InteractionWithResponseBuilder) BinaryBody(body []byte) *V2InteractionWithResponseBuilder {
	i.interaction.interaction.WithBinaryResponseBody(body)

	return i
}

// MultipartBody adds a multipart  body to the expected response
func (i *V2InteractionWithResponseBuilder) MultipartBody(contentType string, filename string, mimePartName string) *V2InteractionWithResponseBuilder {
	i.interaction.interaction.WithResponseMultipartFile(contentType, filename, mimePartName)

	return i
}

// Body adds general body to the expected request
func (i *V2InteractionWithResponseBuilder) Body(contentType string, body []byte) *V2InteractionWithResponseBuilder {
	i.interaction.interaction.WithResponseBody(contentType, body)

	return i
}

// BodyMatch uses struct tags to automatically determine matchers from the given struct
func (i *V2InteractionWithResponseBuilder) BodyMatch(body interface{}) *V2InteractionWithResponseBuilder {
	i.interaction.interaction.WithJSONResponseBody(matchers.MatchV2(body))

	return i
}

// ExecuteTest runs the current test case against a Mock Service.
func (m *V2InteractionWithResponse) ExecuteTest(t *testing.T, integrationTest func(MockServerConfig) error) error {
	return m.provider.ExecuteTest(t, integrationTest)
}
