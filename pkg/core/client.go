package core

import (
	"context"
	"fmt"
)

// Request is the request passed to client Invoke.
type Request interface {
	String() string
}

// Response is the response the client Invoke returns.
type Response interface {
	String() string
}

// RequestGenerator generates a request.
type RequestGenerator interface {
	Generate() Request
}

// Client applies the request to the database.
// Client is used in contorl.
// You should define your own client for your database.
type Client interface {
	// SetUp sets up the client.
	SetUp(ctx context.Context, node string) error
	// TearDown tears down the client.
	TearDown(ctx context.Context, node string) error
	// Invoke invokes a request to the database.
	Invoke(ctx context.Context, node string, r Request) (Response, error)
}

// ClientCreator creates a client.
// The control will create one client for one node.
type ClientCreator interface {
	// Create creates the client.
	Create(node string) Client
	// CreateRequestGenerator creates the request generator.
	CreateRequestGenerator() RequestGenerator
}

type noopClientCreator struct {
}

// Create creates the client.
func (noopClientCreator) Create(node string) Client {
	return noopClient{}
}

// CreateRequestGenerator creates the request generator.
func (noopClientCreator) CreateRequestGenerator() RequestGenerator {
	return noopRequestGenerator{}
}

// noopClient is a noop client
type noopClient struct {
}

// SetUp sets up the client.
func (noopClient) SetUp(ctx context.Context, node string) error { return nil }

// TearDown tears down the client.
func (noopClient) TearDown(ctx context.Context, node string) error { return nil }

// Invoke invokes a request to the database.
func (noopClient) Invoke(ctx context.Context, node string, r Request) (Response, error) {
	return noopResponse{}, nil
}

type noopRequest struct{}

func (noopRequest) String() string {
	return "ok"
}

type noopResponse struct{}

func (noopResponse) String() string {
	return "ok"
}

// noopRequestGenerator generates noop request.
type noopRequestGenerator struct {
}

// Generate implementes Generate interface.
func (noopRequestGenerator) Generate() Request {
	return noopRequest{}
}

var clientCreators = map[string]ClientCreator{}

// RegisterClientCreator registers the client creator for the associated db. Not thread-safe
func RegisterClientCreator(db string, c ClientCreator) {
	_, ok := clientCreators[db]
	if ok {
		panic(fmt.Sprintf("client creator %s is already registered", db))
	}

	clientCreators[db] = c
}

// GetClientCreator gets the registered client creator.
func GetClientCreator(name string) ClientCreator {
	return clientCreators[name]
}

func init() {
	RegisterClientCreator("noop", noopClientCreator{})
}
