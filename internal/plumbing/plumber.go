package plumbing

import "github.com/abjrcode/swervo/internal/app"

type SinkInstance struct {
	SinkCode string `json:"sinkCode"`
	SinkId   string `json:"sinkId"`
}

type DisconnectSinkCommandInput struct {
	SinkCode string `json:"sinkCode"`
	SinkId   string `json:"sinkId"`
}

type Plumber[T interface{}] interface {
	SinkCode() string

	ListConnectedSinks(ctx app.Context, providerCode, providerId string) ([]SinkInstance, error)

	DisconnectSink(ctx app.Context, input DisconnectSinkCommandInput) error

	FlowData(ctx app.Context, data T, sinkId string) error
}
