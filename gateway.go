package theta

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/apex/gateway"
	"github.com/aws/aws-lambda-go/events"
	"github.com/go-chi/chi/v5"
	"github.com/phogolabs/log"
)

// GatewayRoute represents a gateway route
type GatewayRoute interface {
	Mount(r chi.Router)
}

// GatewayInterceptor represents a gateway interceptor
type GatewayInterceptor func(http.Handler) http.Handler

// GatewayHandler represents a webhook
type GatewayHandler struct {
	router chi.Router
}

// Use use a given interceptor.
func (h *GatewayHandler) Use(m GatewayInterceptor) {
	h.mux().Use(m)
}

// Group returns the router.
func (h *GatewayHandler) Mount(route GatewayRoute) {
	route.Mount(h.mux())
}

// HandleContext handles the API Gateway Proxy Request
func (h *GatewayHandler) HandleContext(ctx context.Context, r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	logger := log.WithFields(
		log.Map{
			"gateway_http_method":                   r.HTTPMethod,
			"gateway_path":                          r.Path,
			"gateway_resource":                      r.Resource,
			"gateway_request_context_resource_id":   r.RequestContext.ResourceID,
			"gateway_request_context_resource_path": r.RequestContext.ResourcePath,
			"gateway_request_context_request_id":    r.RequestContext.RequestID,
			"gateway_request_context_stage":         r.RequestContext.Stage,
		},
	)
	// enrich the context
	ctx = log.SetContext(ctx, logger)

	logger.Info("prepare a HTTP request")
	// prepare the request
	request, err := gateway.NewRequest(ctx, r)
	// suppress the execution
	if err != nil {
		logger.WithError(err).Error("prepare a HTTP request failure")
		return events.APIGatewayProxyResponse{}, err
	}

	logger.Info("serve a HTTP request")
	// prepare the response
	response := gateway.NewResponse()
	// delegate to the router
	h.mux().ServeHTTP(response, request)
	// end
	return response.End(), nil
}

func (h *GatewayHandler) mux() chi.Router {
	if h.router == nil {
		h.router = chi.NewMux()
	}

	return h.router
}

// AuthConfig represents an auth configuration.
type GatewayHandlerAuthConfig struct {
	Users []*GatewayHandlerAuthUserConfig
}

// GatewayHandlerAuthUserConfig represents an user configuration.
type GatewayHandlerAuthUserConfig struct {
	Name     string
	Password string
}

// GatewayHandlerAuth represents a authentication handler
type GatewayHandlerAuth struct {
	// Credentials keeps a map of the registered credentials
	Config *GatewayHandlerAuthConfig
}

// HandleContext handles the authorization request
func (p *GatewayHandlerAuth) HandleContext(ctx context.Context, r events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	var (
		unauthorized = fmt.Errorf("Unauthorized")
		empty        = events.APIGatewayCustomAuthorizerResponse{}
	)

	logger := log.WithFields(
		log.Map{
			"gateway_auth_method_arn": r.MethodArn,
			"gateway_auth_type":       r.Type,
		},
	)

	tUsername, tPassword, ok := p.credentials(r.AuthorizationToken)
	if !ok {
		logger.Info("no credentials provided")
		// suppress
		return empty, unauthorized
	}

	logger.Info("token authorization")

	for _, user := range p.Config.Users {
		if user.Name == tUsername && user.Password == tPassword {
			logger.Info("token authorizion succeeded")
			// create a policy
			return p.policy(user.Name, "Allow", "*"), nil
		}
	}

	logger.Info("token authorizion failure")
	// suppress
	return empty, unauthorized
}

func (h *GatewayHandlerAuth) credentials(auth string) (username, password string, ok bool) {
	const prefix = "Basic "

	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return
	}

	data, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}

	var (
		parts = string(data)
		index = strings.IndexByte(parts, ':')
	)

	if index < 0 {
		return
	}

	return parts[:index], parts[index+1:], true
}

func (h *GatewayHandlerAuth) policy(principal, effect, resource string) events.APIGatewayCustomAuthorizerResponse {
	response := events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: principal,
	}

	if effect != "" && resource != "" {
		response.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Resource: []string{resource},
					Effect:   effect,
				},
			},
		}
	}

	response.Context = map[string]interface{}{
		"principal_id": principal,
	}

	return response
}
