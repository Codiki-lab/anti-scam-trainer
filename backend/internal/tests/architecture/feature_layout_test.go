package architecture_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestProductionFeaturesFollowADR0012 keeps the production composition graph
// aligned with the four required anchors of ADR-0012. It deliberately checks
// only the public composition seam and feature entry points.
func TestProductionFeaturesFollowADR0012(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	entries, err := os.ReadDir(filepath.Join(root, "internal/features"))
	if err != nil {
		t.Fatal(err)
	}
	features := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			features = append(features, entry.Name())
		}
	}
	for _, feature := range features {
		for _, relative := range []string{
			"internal/features/" + feature + "/transport/http/handler.go",
			"internal/features/" + feature + "/service/service.go",
			"internal/features/" + feature + "/service/repository.go",
			"internal/features/" + feature + "/repository/postgres.go",
		} {
			if _, err := os.Stat(filepath.Join(root, relative)); err != nil {
				t.Errorf("%s must exist: %v", relative, err)
			}
		}
		assertOnlyADR0012Anchors(t, root, feature)
	}
	for _, removed := range []string{"progress", "users"} {
		if _, err := os.Stat(filepath.Join(root, "internal/features", removed)); !os.IsNotExist(err) {
			t.Errorf("removed feature %q remains in the production tree", removed)
		}
	}
	for _, legacy := range []string{
		"internal/features/attempts/transport/http/game.go",
		"internal/features/scenarios/transport/http/admin.go",
	} {
		if _, err := os.Stat(filepath.Join(root, legacy)); !os.IsNotExist(err) {
			t.Errorf("legacy transport adapter %q remains alongside the active handler", legacy)
		}
	}

	composition, err := os.ReadFile(filepath.Join(root, "internal/core/app/app.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range []string{
		"authhttp.NewWithRateLimits(authentication",
		"learninghttp.New(learning).Routes()",
		"scenarioshttp.New(content).Routes()",
		"attemptshttp.New(game).Routes()",
	} {
		if !strings.Contains(string(composition), route) {
			t.Errorf("composition root does not register active feature route %q", route)
		}
	}
	assertOpenAPIPathsMatchProductionRoutes(t, root, string(composition), features)
	assertLegacyImplementationsAreAbsent(t, root)
}

func assertOnlyADR0012Anchors(t *testing.T, root, feature string) {
	t.Helper()
	want := map[string]bool{
		"repository/postgres.go":    true,
		"service/repository.go":     true,
		"service/service.go":        true,
		"transport/http/handler.go": true,
	}
	// ADR-0014 gives the attempts feature two independently changing AI seams:
	// the typed external port/validation and the compact policy catalogue.
	if feature == "attempts" {
		want["service/ai.go"] = true
		want["service/ai_policy.go"] = true
	}
	featureRoot := filepath.Join(root, "internal/features", feature)
	err := filepath.WalkDir(featureRoot, func(file string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(file) != ".go" {
			return nil
		}
		relative, err := filepath.Rel(featureRoot, file)
		if err != nil {
			return err
		}
		if !want[filepath.ToSlash(relative)] {
			t.Errorf("feature %q retains non-ADR-0012 file %q without an explicit architectural exception", feature, relative)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func assertOpenAPIPathsMatchProductionRoutes(t *testing.T, root, composition string, features []string) {
	t.Helper()
	specification, err := os.ReadFile(filepath.Join(root, "openapi/v1/openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	pathExpression := regexp.MustCompile(`(?m)^  (/api/v1/[^:]+):$`)
	openAPIPaths := pathExpression.FindAllStringSubmatch(string(specification), -1)
	registered := []string{"/health"}
	routeExpression := regexp.MustCompile(`Path:\s*"([^"]+)"`)
	for _, feature := range features {
		packageAlias := feature + "http"
		if feature == "auth" {
			packageAlias = "authhttp"
		}
		if !strings.Contains(composition, packageAlias+".") {
			t.Errorf("feature %q is not imported and registered by the production composition root", feature)
			continue
		}
		files, globErr := filepath.Glob(filepath.Join(root, "internal/features", feature, "transport/http/*.go"))
		if globErr != nil {
			t.Fatal(globErr)
		}
		for _, file := range files {
			content, readErr := os.ReadFile(file)
			if readErr != nil {
				t.Fatal(readErr)
			}
			name := filepath.Base(file)
			active := name == "handler.go" || feature == "learning" && name == "admin.go"
			if !active {
				if strings.Contains(string(content), "Routes() []router.Route") {
					t.Errorf("unregistered parallel transport adapter %q defines routes", file)
				}
				continue
			}
			for _, match := range routeExpression.FindAllStringSubmatch(string(content), -1) {
				registered = append(registered, match[1])
			}
		}
	}
	for _, match := range openAPIPaths {
		path := strings.TrimPrefix(match[1], "/api/v1")
		if !matchesRegisteredRoute(path, registered) {
			t.Errorf("OpenAPI v1 path %q has no route in an active production feature handler", match[1])
		}
	}
}

func matchesRegisteredRoute(path string, routes []string) bool {
	for _, route := range routes {
		if route == path || strings.HasSuffix(route, "/") && strings.HasPrefix(path, route) {
			return true
		}
	}
	return false
}

func assertLegacyImplementationsAreAbsent(t *testing.T, root string) {
	t.Helper()
	for file, forbidden := range map[string][]string{
		"internal/features/attempts/service/service.go":  {"type Service struct", "func New(repository Repository"},
		"internal/features/scenarios/service/service.go": {"type ContentService", "func NewContent("},
	} {
		content, err := os.ReadFile(filepath.Join(root, file))
		if err != nil {
			t.Fatal(err)
		}
		for _, symbol := range forbidden {
			if strings.Contains(string(content), symbol) {
				t.Errorf("legacy implementation %q remains in %s", symbol, file)
			}
		}
	}
}
