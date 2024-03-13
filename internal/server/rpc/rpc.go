package rpc

import (
	"context"

	"github.com/Jourloy/go-metrics-collector/internal/proto"
	"github.com/Jourloy/go-metrics-collector/internal/server/storage"
)

type MetricServer struct {
	proto.UnimplementedMetricServiceServer
	storage storage.Storage
}

func (s *MetricServer) UpdateCounter(ctx context.Context, in *proto.UpdateCounterRequest) (*proto.UpdateResponse, error) {
	var response proto.UpdateResponse

	// Update metric
	s.storage.UpdateCounterMetric(in.Name, in.Value)

	return &response, nil
}

func (s *MetricServer) UpdateGauge(ctx context.Context, in *proto.UpdateGaugeRequest) (*proto.UpdateResponse, error) {
	var response proto.UpdateResponse

	// Update metric
	s.storage.UpdateGaugeMetric(in.Name, in.Value)

	return &response, nil
}
