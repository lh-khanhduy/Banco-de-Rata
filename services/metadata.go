package services

import (
	"context"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	userAgentHeader            = "user-agent"
	xForwardedForHeader        = "x-forwarded-for"
)

type Metadata struct {
	UserAgent string
	ClientIP  string
}

func (s *Server) extractMetadata(ctx context.Context) *Metadata {
	result := &Metadata{}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// HTTP GATEWAY
		if userAgents := md.Get(grpcGatewayUserAgentHeader); len(userAgents) > 0 {
			result.UserAgent = userAgents[0]
		}

		if clientIPs := md.Get(xForwardedForHeader); len(clientIPs) > 0 {
			result.ClientIP = clientIPs[0]
		}

		// gRPC
		if userAgents := md.Get(userAgentHeader); len(userAgents) > 0 {
			result.UserAgent = userAgents[0]
		}
	}

	if p, ok := peer.FromContext(ctx); ok {
		result.ClientIP = p.Addr.String()
	}

	return result
}
