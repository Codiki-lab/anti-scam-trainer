package app

import (
	"anti-scam-trainer/backend/internal/core/config"
	"anti-scam-trainer/backend/internal/core/logger"
	"anti-scam-trainer/backend/internal/core/server/middleware"
	"anti-scam-trainer/backend/internal/core/server/openapidocs"
	"anti-scam-trainer/backend/internal/core/server/response"
	"anti-scam-trainer/backend/internal/core/server/router"
	attemptshttp "anti-scam-trainer/backend/internal/features/attempts/transport/http"
	authhttp "anti-scam-trainer/backend/internal/features/auth/transport/http"
	learninghttp "anti-scam-trainer/backend/internal/features/learning/transport/http"
	scenarioshttp "anti-scam-trainer/backend/internal/features/scenarios/transport/http"
	"net/http"
)

func newHandler(cfg config.Config, log *logger.Logger, dependencies dependencies) http.Handler {
	versionedRouter := router.New()
	versionedRouter.Register(router.V1, []router.Route{{Path: "/health", Handler: health}})
	versionedRouter.Register(router.V1, authhttp.NewWithRateLimits(dependencies.authentication, dependencies.registration, dependencies.login, dependencies.clientIP).Routes())
	versionedRouter.Register(router.V1, learninghttp.NewWithChatRecommendation(dependencies.learning, dependencies.chatRecommendation).Routes())
	versionedRouter.Register(router.V1, learninghttp.NewAdmin(dependencies.learningContent).Routes())
	versionedRouter.Register(router.V1, scenarioshttp.New(dependencies.content).Routes())
	versionedRouter.Register(router.V1, attemptshttp.New(dependencies.game).Routes())

	routes := http.NewServeMux()
	documentationHandler := middleware.RequireSwaggerAuthentication(cfg.SwaggerUsername, cfg.SwaggerPassword)(openapidocs.NewHandler())
	routes.Handle("/swagger", http.RedirectHandler("/swagger/", http.StatusTemporaryRedirect))
	routes.Handle("/swagger/", documentationHandler)
	routes.Handle("/openapi/", documentationHandler)
	routes.Handle("/", versionedRouter)
	return middleware.Chain(routes, middleware.RequestID(), middleware.CORS(cfg.FrontendOrigins), middleware.Logger(log), middleware.Panic(), middleware.Trace(), authhttp.RequireAuthentication(dependencies.tokens))
}

func health(writer http.ResponseWriter, _ *http.Request) {
	response.JSON(writer, map[string]string{"status": "ok"})
}
