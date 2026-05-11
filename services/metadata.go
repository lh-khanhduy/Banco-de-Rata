package services

import (
	"context"
	"encoding/json"
	"log"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

const (
	grpcGatewayUserAgentHeader = "grpcgateway-user-agent"
	userAgentHeader            = "user-agent"
	xForwardedForHeader        = "x-forwarded-for"
	authorizationPayloadKey    = "authorization_payload"
)

type Metadata struct {
	Username  string
	UserAgent string
	ClientIP  string
}

type AuthPayload struct {
	Username string `json:"username"`
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

		// AuthPayload extraction
		if payloads := md.Get(authorizationPayloadKey); len(payloads) > 0 {
			var authPayload AuthPayload
			if err := json.Unmarshal([]byte(payloads[0]), &authPayload); err == nil {
				result.Username = authPayload.Username
			} else {
				// Optional: Log the error if you want visibility into bad payloads
				log.Printf("failed to unmarshal authorization_payload: %v", err)
			}
		}
	}

	if p, ok := peer.FromContext(ctx); ok {
		result.ClientIP = p.Addr.String()
	}

	return result
}
