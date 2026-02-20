// Command generate creates starter .graft files for common development domains.
// Each graft bundles curated, high-value memory seeds that teams can import
// via `phloem graft import <file.graft>`.
//
// Usage:
//
//	go run ./phloem/grafts/generate
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/CanopyHQ/phloem/internal/graft"
	"github.com/CanopyHQ/phloem/internal/memory"
)

func main() {
	// Determine output directory (same as this file's parent, i.e. phloem/grafts/)
	outputDir := filepath.Join(".", "phloem", "grafts")

	// If run from the phloem directory itself, adjust
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		// Try relative to current working directory
		cwd, _ := os.Getwd()
		// Check if we are inside phloem/grafts/generate
		if filepath.Base(cwd) == "generate" {
			outputDir = filepath.Dir(cwd)
		} else if filepath.Base(cwd) == "grafts" {
			outputDir = cwd
		} else if filepath.Base(cwd) == "phloem" {
			outputDir = filepath.Join(cwd, "grafts")
		} else {
			// Fallback: write to current directory
			outputDir = cwd
		}
	}

	grafts := []struct {
		filename string
		manifest graft.Manifest
		memories []memory.Memory
	}{
		{
			filename: "go-best-practices.graft",
			manifest: graft.Manifest{
				ID:          "go-best-practices-v1",
				Name:        "Go Best Practices",
				Description: "Curated memory seeds for idiomatic Go development: error handling, testing, concurrency, and API design patterns.",
				Author:      "Canopy Team",
				Version:     "1.0.0",
				CreatedAt:   time.Now().UTC(),
				MemoryCount: len(goBestPractices()),
				Tags:        []string{"go", "best-practices", "starter"},
			},
			memories: goBestPractices(),
		},
		{
			filename: "react-patterns.graft",
			manifest: graft.Manifest{
				ID:          "react-patterns-v1",
				Name:        "React Patterns",
				Description: "Curated memory seeds for modern React development: hooks, component patterns, state management, and testing.",
				Author:      "Canopy Team",
				Version:     "1.0.0",
				CreatedAt:   time.Now().UTC(),
				MemoryCount: len(reactPatterns()),
				Tags:        []string{"react", "typescript", "frontend", "starter"},
			},
			memories: reactPatterns(),
		},
		{
			filename: "security-essentials.graft",
			manifest: graft.Manifest{
				ID:          "security-essentials-v1",
				Name:        "Security Essentials",
				Description: "Curated memory seeds for application security: OWASP Top 10, authentication, authorization, and secure coding practices.",
				Author:      "Canopy Team",
				Version:     "1.0.0",
				CreatedAt:   time.Now().UTC(),
				MemoryCount: len(securityEssentials()),
				Tags:        []string{"security", "owasp", "best-practices", "starter"},
			},
			memories: securityEssentials(),
		},
	}

	for _, g := range grafts {
		outPath := filepath.Join(outputDir, g.filename)
		if err := graft.Package(g.manifest, g.memories, nil, outPath); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: failed to create %s: %v\n", g.filename, err)
			os.Exit(1)
		}
		fmt.Printf("Created %s (%d memories)\n", outPath, len(g.memories))
	}

	fmt.Println("\nDone. Import with: phloem graft import <file.graft>")
}

// helper creates a Memory with sensible defaults.
func mem(id, content, context string, tags []string) memory.Memory {
	now := time.Now().UTC()
	return memory.Memory{
		ID:        id,
		Content:   content,
		Context:   context,
		Tags:      tags,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ---------------------------------------------------------------------------
// Go Best Practices
// ---------------------------------------------------------------------------

func goBestPractices() []memory.Memory {
	return []memory.Memory{
		mem(
			"go-error-wrapping",
			"Always wrap errors with context using fmt.Errorf and the %w verb so callers can unwrap the chain with errors.Is and errors.As. A bare `return err` loses the call-site context and makes debugging painful. Pattern: `return fmt.Errorf(\"loading config: %w\", err)`.",
			"Go error handling",
			[]string{"go", "error-handling", "best-practice"},
		),
		mem(
			"go-sentinel-errors",
			"Define sentinel errors as package-level variables with errors.New for conditions callers need to check programmatically (e.g., `var ErrNotFound = errors.New(\"not found\")`). Use custom error types when you need to carry extra data. Check with errors.Is for sentinels and errors.As for typed errors.",
			"Go error handling",
			[]string{"go", "error-handling", "sentinel-errors", "best-practice"},
		),
		mem(
			"go-context-usage",
			"Pass context.Context as the first parameter to any function that does I/O or may block. Use context.WithTimeout for external calls (HTTP, DB, gRPC) to prevent resource leaks. Never store contexts in structs; they belong on the call path. Respect ctx.Done() in long-running loops.",
			"Go context patterns",
			[]string{"go", "context", "concurrency", "best-practice"},
		),
		mem(
			"go-table-driven-tests",
			"Use table-driven tests for any function with multiple input/output cases. Define a slice of anonymous structs with `name`, input fields, and `want` fields, then range over them calling t.Run(tc.name, ...). This keeps tests readable, easy to extend, and gives clear failure messages per case.",
			"Go testing patterns",
			[]string{"go", "testing", "table-driven", "best-practice"},
		),
		mem(
			"go-interface-segregation",
			"Accept interfaces, return concrete structs. Keep interfaces small (1-3 methods) and define them in the consuming package, not the implementing package. This follows the Interface Segregation Principle and makes mocking in tests trivial. The standard library's io.Reader and io.Writer are the gold standard.",
			"Go interface design",
			[]string{"go", "interfaces", "design", "best-practice"},
		),
		mem(
			"go-goroutine-lifecycle",
			"Every goroutine you start must have a clear shutdown path. Use context cancellation or a done channel to signal goroutines to exit. For groups of goroutines, use errgroup.Group which handles cancellation and error propagation. A goroutine leak is as serious as a memory leak.",
			"Go concurrency",
			[]string{"go", "goroutines", "concurrency", "errgroup", "best-practice"},
		),
		mem(
			"go-dependency-injection",
			"Pass dependencies as explicit constructor parameters rather than relying on globals or init(). Define a NewService(deps...) constructor that returns a struct holding its dependencies. This makes code testable (swap in mocks), readable (dependencies are visible), and avoids hidden coupling.",
			"Go dependency injection",
			[]string{"go", "dependency-injection", "design", "testing", "best-practice"},
		),
		mem(
			"go-package-naming",
			"Package names should be short, lowercase, single-word nouns (e.g., `http`, `user`, `store`). Avoid generic names like `util`, `common`, or `helpers` -- they become junk drawers. The package name is part of the caller's vocabulary: `user.New()` reads better than `models.NewUser()`. Never stutter: `http.HTTPClient` is wrong; `http.Client` is right.",
			"Go package design",
			[]string{"go", "packages", "naming", "best-practice"},
		),
		mem(
			"go-functional-options",
			"Use the functional options pattern for constructors with many optional parameters. Define `type Option func(*Config)` and provide option functions like `WithTimeout(d time.Duration) Option`. This avoids long parameter lists, config struct sprawl, and the boolean-flag trap. It also lets you add options without breaking callers.",
			"Go struct initialization",
			[]string{"go", "functional-options", "constructor", "api-design", "best-practice"},
		),
		mem(
			"go-module-versioning",
			"Follow semantic versioning for Go modules. For v0/v1, the import path is just the module path. For v2+, append /v2 to the module path in both go.mod and import statements. Use go mod tidy regularly. Pin dependencies in production and review go.sum changes in code review to catch supply-chain issues.",
			"Go modules",
			[]string{"go", "modules", "versioning", "best-practice"},
		),
		mem(
			"go-structured-logging",
			"Use log/slog (standard library since Go 1.21) for structured logging. Create a logger with slog.New(slog.NewJSONHandler(os.Stdout, nil)) for production and slog.NewTextHandler for development. Always include contextual key-value pairs: slog.Info(\"request handled\", \"method\", r.Method, \"path\", r.URL.Path, \"duration\", elapsed). Avoid string interpolation in log messages.",
			"Go logging",
			[]string{"go", "logging", "slog", "observability", "best-practice"},
		),
		mem(
			"go-error-types",
			"Create custom error types when callers need structured error data beyond a message. Implement the error interface and optionally Unwrap() for wrapping. Example: `type ValidationError struct { Field, Message string }`. Use errors.As in callers to extract the typed error. Reserve custom types for domain-specific errors; use fmt.Errorf with %w for general wrapping.",
			"Go error handling",
			[]string{"go", "error-handling", "custom-errors", "best-practice"},
		),
	}
}

// ---------------------------------------------------------------------------
// React Patterns
// ---------------------------------------------------------------------------

func reactPatterns() []memory.Memory {
	return []memory.Memory{
		mem(
			"react-custom-hooks",
			"Extract reusable stateful logic into custom hooks (functions starting with 'use'). A custom hook can call other hooks and return values or callbacks. Examples: useLocalStorage, useDebounce, useFetch. This is React's primary mechanism for code reuse -- prefer it over render props or HOCs.",
			"React hooks",
			[]string{"react", "hooks", "custom-hooks", "best-practice"},
		),
		mem(
			"react-error-boundaries",
			"Wrap component trees with Error Boundaries (class components implementing getDerivedStateFromError or componentDidCatch) to catch rendering errors gracefully. Without them, a single broken component crashes the entire app. Place boundaries around routes, feature sections, and third-party widgets. Show a fallback UI, not a white screen.",
			"React error handling",
			[]string{"react", "error-boundaries", "error-handling", "best-practice"},
		),
		mem(
			"react-composition",
			"Prefer component composition over inheritance. Use children props and render slots to build flexible layouts. Specialization is done by rendering specific content inside generic containers, not by extending base classes. React has no class hierarchy -- composition is the design model.",
			"React component design",
			[]string{"react", "composition", "component-design", "best-practice"},
		),
		mem(
			"react-state-management",
			"Choose state tools by scope: useState for local component state, useReducer for complex state with multiple sub-values or transitions, and Context (with useContext) for data needed across many components (theme, auth, locale). For server state, use React Query or SWR instead of manual useEffect + useState. Avoid putting everything in global state.",
			"React state management",
			[]string{"react", "state", "useState", "useReducer", "context", "best-practice"},
		),
		mem(
			"react-memoization",
			"Use useMemo for expensive computations and useCallback for stable function references passed to child components. But do not memoize everything -- premature memoization adds complexity with no benefit. Memoize only when you have measured a performance problem or when a reference-stable callback prevents child re-renders in a list.",
			"React performance",
			[]string{"react", "useMemo", "useCallback", "performance", "best-practice"},
		),
		mem(
			"react-key-prop",
			"Never use array index as a key for dynamic lists where items can be added, removed, or reordered. Index keys cause incorrect component reuse, stale state, and subtle UI bugs. Use a stable, unique identifier (database ID, UUID) as the key. Index keys are acceptable only for static, never-reordered lists.",
			"React lists and keys",
			[]string{"react", "keys", "lists", "best-practice"},
		),
		mem(
			"react-controlled-components",
			"Controlled components store their value in React state and update via onChange. Uncontrolled components use refs and the DOM as the source of truth. Prefer controlled components for form inputs that need validation, conditional logic, or cross-field dependencies. Use uncontrolled components (with useRef) only for simple cases or when integrating with non-React code.",
			"React forms",
			[]string{"react", "forms", "controlled-components", "best-practice"},
		),
		mem(
			"react-data-fetching",
			"When fetching data in useEffect, always return a cleanup function that sets an 'ignore' flag or calls AbortController.abort() to prevent setting state on an unmounted component. Better yet, use a data-fetching library (React Query, SWR, or the use() hook with Suspense) that handles caching, deduplication, race conditions, and cleanup automatically.",
			"React data fetching",
			[]string{"react", "data-fetching", "useEffect", "abort-controller", "best-practice"},
		),
		mem(
			"react-typescript",
			"Type component props with an interface, not React.FC (which adds implicit children and has had breaking changes). Use `function MyComponent(props: MyProps)` or destructure directly. For event handlers, use React's built-in event types: React.MouseEvent, React.ChangeEvent<HTMLInputElement>. For generic components, use `<T,>` to avoid JSX ambiguity.",
			"React with TypeScript",
			[]string{"react", "typescript", "types", "best-practice"},
		),
		mem(
			"react-testing-library",
			"Test components by simulating user behavior, not by inspecting implementation details. Query by role, label text, or placeholder -- not by class names or component internals. Use screen.getByRole('button', { name: /submit/i }) instead of finding by test ID. Assert on what the user sees, not on state or props.",
			"React testing",
			[]string{"react", "testing", "react-testing-library", "best-practice"},
		),
		mem(
			"react-effect-cleanup",
			"Every useEffect that subscribes to events, starts timers, or opens connections must return a cleanup function. React calls cleanup before re-running the effect and on unmount. Missing cleanup causes memory leaks, stale listeners, and zombie subscriptions. Think of effects as 'synchronize with X' and cleanup as 'stop synchronizing with X'.",
			"React useEffect",
			[]string{"react", "useEffect", "cleanup", "best-practice"},
		),
		mem(
			"react-component-file-structure",
			"Co-locate component files: keep the component, its styles, its tests, and its types in the same directory. A component directory might contain Button.tsx, Button.test.tsx, Button.module.css, and index.ts. This makes components portable and easy to find. Avoid organizing by file type (all components in /components, all tests in /tests).",
			"React project structure",
			[]string{"react", "project-structure", "organization", "best-practice"},
		),
	}
}

// ---------------------------------------------------------------------------
// Security Essentials
// ---------------------------------------------------------------------------

func securityEssentials() []memory.Memory {
	return []memory.Memory{
		mem(
			"sec-input-validation",
			"Validate all external input at the boundary where it enters your system (HTTP handlers, CLI parsers, message consumers). Use an allowlist approach: define what is valid, reject everything else. Never rely on client-side validation alone. Validate type, length, range, and format. This is the single most effective security control.",
			"Input validation",
			[]string{"security", "input-validation", "boundary", "best-practice"},
		),
		mem(
			"sec-sql-injection",
			"Always use parameterized queries or prepared statements for SQL. Never concatenate user input into SQL strings. This applies to every language and ORM -- even if you think the input is 'safe'. SQL injection remains a top-3 vulnerability decade after decade because developers skip this rule 'just once'.",
			"SQL injection prevention",
			[]string{"security", "sql-injection", "database", "owasp", "best-practice"},
		),
		mem(
			"sec-xss-prevention",
			"Prevent XSS by escaping all dynamic output inserted into HTML, JavaScript, CSS, or URLs. Use your framework's built-in auto-escaping (React does this by default for JSX). Set Content-Security-Policy headers to restrict inline scripts and external sources. Never use dangerouslySetInnerHTML (React) or innerHTML without sanitization.",
			"XSS prevention",
			[]string{"security", "xss", "csp", "owasp", "best-practice"},
		),
		mem(
			"sec-csrf-protection",
			"Protect state-changing endpoints from CSRF attacks. Use SameSite=Lax or SameSite=Strict on cookies (this blocks most CSRF). For additional defense, use anti-CSRF tokens (synchronizer tokens or double-submit cookies). APIs using Authorization headers (Bearer tokens) instead of cookies are inherently CSRF-resistant.",
			"CSRF protection",
			[]string{"security", "csrf", "cookies", "owasp", "best-practice"},
		),
		mem(
			"sec-password-storage",
			"Hash passwords with bcrypt (cost factor 12+) or argon2id. Never store plaintext passwords, MD5, SHA-256, or any fast hash. Use a unique salt per password (bcrypt and argon2 do this automatically). Enforce minimum password length (8+ characters) but avoid overly complex rules that push users toward weaker patterns.",
			"Password storage",
			[]string{"security", "authentication", "passwords", "bcrypt", "argon2", "best-practice"},
		),
		mem(
			"sec-authorization",
			"Check permissions at the handler/controller level, not deep inside business logic. Use role-based access control (RBAC) or attribute-based access control (ABAC) depending on complexity. Always verify that the authenticated user is authorized to access the specific resource (IDOR prevention). Deny by default; grant explicitly.",
			"Authorization patterns",
			[]string{"security", "authorization", "rbac", "idor", "best-practice"},
		),
		mem(
			"sec-secret-management",
			"Never commit secrets (API keys, passwords, tokens) to version control. Use environment variables or a secret manager (Vault, AWS Secrets Manager, 1Password CLI). Add .env files to .gitignore. Rotate secrets regularly. If a secret is accidentally committed, consider it compromised -- rotate it immediately, don't just delete the commit.",
			"Secret management",
			[]string{"security", "secrets", "environment-variables", "gitignore", "best-practice"},
		),
		mem(
			"sec-https-everywhere",
			"Serve all traffic over HTTPS. Set the Strict-Transport-Security (HSTS) header with a long max-age (e.g., 31536000 seconds) and includeSubDomains. Mark cookies as Secure (only sent over HTTPS) and HttpOnly (not accessible to JavaScript). Use TLS 1.2+ and disable older protocols.",
			"HTTPS and transport security",
			[]string{"security", "https", "tls", "hsts", "cookies", "best-practice"},
		),
		mem(
			"sec-rate-limiting",
			"Apply rate limiting to authentication endpoints, API routes, and any expensive operations. Use token bucket or sliding window algorithms. Return HTTP 429 with a Retry-After header. Rate limit by IP for anonymous endpoints and by user ID for authenticated ones. This protects against brute force, credential stuffing, and denial-of-service.",
			"Rate limiting",
			[]string{"security", "rate-limiting", "dos", "brute-force", "best-practice"},
		),
		mem(
			"sec-dependency-scanning",
			"Keep dependencies updated and scan them continuously for known vulnerabilities. Use tools like Dependabot, Snyk, or govulncheck (Go). Pin dependency versions in production. Review lock file changes in code review. A single vulnerable transitive dependency can compromise your entire application.",
			"Dependency security",
			[]string{"security", "dependencies", "supply-chain", "scanning", "best-practice"},
		),
		mem(
			"sec-jwt-best-practices",
			"Keep JWT access tokens short-lived (5-15 minutes) and use refresh tokens for longer sessions. Always validate the signature, issuer (iss), audience (aud), and expiration (exp) claims. Use asymmetric signing (RS256/ES256) for distributed systems. Never store sensitive data in JWT payloads -- they are base64-encoded, not encrypted.",
			"JWT security",
			[]string{"security", "jwt", "authentication", "tokens", "best-practice"},
		),
		mem(
			"sec-owasp-top-10",
			"The OWASP Top 10 is the baseline security awareness checklist: Broken Access Control, Cryptographic Failures, Injection, Insecure Design, Security Misconfiguration, Vulnerable Components, Authentication Failures, Data Integrity Failures, Logging Failures, and SSRF. Review your application against this list at least once per release cycle. Most breaches exploit these well-known categories.",
			"OWASP Top 10",
			[]string{"security", "owasp", "checklist", "awareness", "best-practice"},
		),
		mem(
			"sec-logging-security",
			"Log security-relevant events: authentication attempts (success and failure), authorization failures, input validation failures, and administrative actions. Never log sensitive data (passwords, tokens, PII). Use structured logging with correlation IDs. Ensure logs are tamper-resistant and retained long enough for incident investigation (90+ days).",
			"Security logging",
			[]string{"security", "logging", "monitoring", "incident-response", "best-practice"},
		),
	}
}
