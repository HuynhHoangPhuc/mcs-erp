package agent

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/application/services"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/delivery"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/agent/infrastructure"
	coreservices "github.com/HuynhHoangPhuc/mcs-erp/internal/core/application/services"
	coredelivery "github.com/HuynhHoangPhuc/mcs-erp/internal/core/delivery"
	coredomain "github.com/HuynhHoangPhuc/mcs-erp/internal/core/domain"
	"github.com/HuynhHoangPhuc/mcs-erp/internal/platform/auth"
)

// Module implements pkg/module.Module for the AI Agent bounded context.
type Module struct {
	pool        *pgxpool.Pool
	authSvc     *coreservices.AuthService
	registry    *infrastructure.ToolRegistry
	providerSvc *services.ProviderService
	convRepo    domain.ConversationRepository
	cache       *infrastructure.RedisMessageCache
	agentSvc    *services.AgentService
}

// NewModule wires all agent dependencies.
// registry is shared with other modules so they can register tools at startup.
func NewModule(
	pool *pgxpool.Pool,
	authSvc *coreservices.AuthService,
	registry *infrastructure.ToolRegistry,
	llmCfg domain.LLMConfig,
	redisURL string,
) *Module {
	convRepo := infrastructure.NewPostgresConversationRepo(pool)
	providerSvc := services.NewProviderService(llmCfg)

	var cache *infrastructure.RedisMessageCache
	if redisURL != "" {
		if c, err := infrastructure.NewRedisMessageCache(redisURL); err == nil {
			cache = c
		}
	}

	agentSvc := services.NewAgentService(convRepo, cache, registry, providerSvc)

	return &Module{
		pool:        pool,
		authSvc:     authSvc,
		registry:    registry,
		providerSvc: providerSvc,
		convRepo:    convRepo,
		cache:       cache,
		agentSvc:    agentSvc,
	}
}

// ToolRegistry exposes the registry so other modules can register tools during bootstrap.
func (m *Module) ToolRegistry() *infrastructure.ToolRegistry { return m.registry }

func (m *Module) Name() string           { return "agent" }
func (m *Module) Dependencies() []string { return []string{"core"} }
func (m *Module) Migrate(_ context.Context) error        { return nil }
func (m *Module) RegisterEvents(_ context.Context) error { return nil }

func (m *Module) RegisterRoutes(mux *http.ServeMux) {
	chatHandler := delivery.NewChatHandler(m.agentSvc, m.convRepo)
	convHandler := delivery.NewConversationHandler(m.convRepo)
	suggHandler := delivery.NewSuggestionHandler()

	authMw := coredelivery.AuthMiddleware(m.authSvc)
	requireChat := auth.RequirePermission(coredomain.PermAgentChat)

	// Chat (SSE streaming)
	mux.Handle("POST /api/v1/agent/chat",
		authMw(requireChat(http.HandlerFunc(chatHandler.HandleChat))))

	// Conversations CRUD
	mux.Handle("GET /api/v1/agent/conversations",
		authMw(requireChat(http.HandlerFunc(convHandler.ListConversations))))
	mux.Handle("POST /api/v1/agent/conversations",
		authMw(requireChat(http.HandlerFunc(convHandler.CreateConversation))))
	mux.Handle("GET /api/v1/agent/conversations/{id}",
		authMw(requireChat(http.HandlerFunc(convHandler.GetConversation))))
	mux.Handle("PATCH /api/v1/agent/conversations/{id}",
		authMw(requireChat(http.HandlerFunc(convHandler.UpdateConversation))))
	mux.Handle("DELETE /api/v1/agent/conversations/{id}",
		authMw(requireChat(http.HandlerFunc(convHandler.DeleteConversation))))

	// Inline suggestions (rule-based, no LLM call)
	mux.Handle("GET /api/v1/agent/suggestions",
		authMw(requireChat(http.HandlerFunc(suggHandler.HandleSuggestions))))
}
