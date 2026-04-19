package skill

import "strings"

// KnownLibrary describes a library in the markdocs registry.
type KnownLibrary struct {
	Name     string
	Category string
	DocHints []string // suggested doc URL prefixes for hinting
}

// Registry is the built-in list of commonly-documented libraries.
// Used by `markdocs scan` to cross-reference detected dependencies.
var Registry = []KnownLibrary{
	// Frontend frameworks
	{Name: "react", Category: "frontend", DocHints: []string{"https://react.dev"}},
	{Name: "vue", Category: "frontend", DocHints: []string{"https://vuejs.org/guide"}},
	{Name: "next", Category: "frontend", DocHints: []string{"https://nextjs.org/docs"}},
	{Name: "nuxt", Category: "frontend", DocHints: []string{"https://nuxt.com/docs"}},
	{Name: "svelte", Category: "frontend", DocHints: []string{"https://svelte.dev/docs"}},
	{Name: "astro", Category: "frontend", DocHints: []string{"https://docs.astro.build"}},
	{Name: "remix", Category: "frontend", DocHints: []string{"https://remix.run/docs"}},
	{Name: "@angular/core", Category: "frontend", DocHints: []string{"https://angular.io/docs"}},
	{Name: "solid-js", Category: "frontend", DocHints: []string{"https://docs.solidjs.com"}},
	{Name: "qwik", Category: "frontend", DocHints: []string{"https://qwik.dev/docs"}},
	// State management
	{Name: "zustand", Category: "frontend", DocHints: []string{"https://zustand.docs.pmnd.rs"}},
	{Name: "jotai", Category: "frontend", DocHints: []string{"https://jotai.org/docs"}},
	{Name: "recoil", Category: "frontend", DocHints: []string{"https://recoiljs.org/docs"}},
	{Name: "redux", Category: "frontend", DocHints: []string{"https://redux.js.org/introduction"}},
	{Name: "@reduxjs/toolkit", Category: "frontend", DocHints: []string{"https://redux-toolkit.js.org/introduction"}},
	{Name: "mobx", Category: "frontend", DocHints: []string{"https://mobx.js.org/README.html"}},
	// UI Libraries
	{Name: "shadcn-ui", Category: "ui", DocHints: []string{"https://ui.shadcn.com/docs"}},
	{Name: "@radix-ui/react-dialog", Category: "ui", DocHints: []string{"https://www.radix-ui.com/docs"}},
	{Name: "tailwindcss", Category: "ui", DocHints: []string{"https://tailwindcss.com/docs"}},
	{Name: "@mui/material", Category: "ui", DocHints: []string{"https://mui.com/material-ui"}},
	{Name: "chakra-ui", Category: "ui", DocHints: []string{"https://chakra-ui.com/docs"}},
	{Name: "antd", Category: "ui", DocHints: []string{"https://ant.design/docs"}},
	{Name: "mantine", Category: "ui", DocHints: []string{"https://mantine.dev/docs"}},
	{Name: "daisyui", Category: "ui", DocHints: []string{"https://daisyui.com/docs"}},
	// Data fetching / async
	{Name: "@tanstack/react-query", Category: "frontend", DocHints: []string{"https://tanstack.com/query/latest/docs"}},
	{Name: "swr", Category: "frontend", DocHints: []string{"https://swr.vercel.app/docs"}},
	{Name: "axios", Category: "backend", DocHints: []string{"https://axios-http.com/docs"}},
	// Backend (JS/TS)
	{Name: "express", Category: "backend", DocHints: []string{"https://expressjs.com"}},
	{Name: "fastify", Category: "backend", DocHints: []string{"https://fastify.dev/docs"}},
	{Name: "koa", Category: "backend", DocHints: []string{"https://koajs.com"}},
	{Name: "hono", Category: "backend", DocHints: []string{"https://hono.dev/docs"}},
	{Name: "nestjs", Category: "backend", DocHints: []string{"https://docs.nestjs.com"}},
	{Name: "trpc", Category: "backend", DocHints: []string{"https://trpc.io/docs"}},
	// Backend (Go)
	{Name: "gin", Category: "backend", DocHints: []string{"https://gin-gonic.com/docs"}},
	{Name: "fiber", Category: "backend", DocHints: []string{"https://docs.gofiber.io"}},
	{Name: "echo", Category: "backend", DocHints: []string{"https://echo.labstack.com/docs"}},
	{Name: "chi", Category: "backend", DocHints: []string{"https://go-chi.io/docs"}},
	// Database
	{Name: "prisma", Category: "database", DocHints: []string{"https://www.prisma.io/docs"}},
	{Name: "drizzle-orm", Category: "database", DocHints: []string{"https://orm.drizzle.team/docs"}},
	{Name: "@supabase/supabase-js", Category: "database", DocHints: []string{"https://supabase.com/docs"}},
	{Name: "mongoose", Category: "database", DocHints: []string{"https://mongoosejs.com/docs"}},
	{Name: "typeorm", Category: "database", DocHints: []string{"https://typeorm.io"}},
	{Name: "kysely", Category: "database", DocHints: []string{"https://kysely.dev/docs"}},
	{Name: "sqlalchemy", Category: "database", DocHints: []string{"https://docs.sqlalchemy.org"}},
	// Testing
	{Name: "vitest", Category: "testing", DocHints: []string{"https://vitest.dev/guide"}},
	{Name: "jest", Category: "testing", DocHints: []string{"https://jestjs.io/docs"}},
	{Name: "playwright", Category: "testing", DocHints: []string{"https://playwright.dev/docs"}},
	{Name: "cypress", Category: "testing", DocHints: []string{"https://docs.cypress.io"}},
	{Name: "@testing-library/react", Category: "testing", DocHints: []string{"https://testing-library.com/docs"}},
	{Name: "pytest", Category: "testing", DocHints: []string{"https://docs.pytest.org"}},
	// Auth
	{Name: "next-auth", Category: "auth", DocHints: []string{"https://next-auth.js.org/getting-started"}},
	{Name: "@auth/core", Category: "auth", DocHints: []string{"https://authjs.dev/getting-started"}},
	{Name: "clerk", Category: "auth", DocHints: []string{"https://clerk.com/docs"}},
	{Name: "lucia", Category: "auth", DocHints: []string{"https://lucia-auth.com/docs"}},
	// Payments
	{Name: "stripe", Category: "payments", DocHints: []string{"https://stripe.com/docs"}},
	{Name: "@lemonsqueezy/lemonsqueezy.js", Category: "payments", DocHints: []string{"https://docs.lemonsqueezy.com"}},
	// Infra / DevOps
	{Name: "docker", Category: "infra", DocHints: []string{"https://docs.docker.com"}},
	{Name: "kubernetes", Category: "infra", DocHints: []string{"https://kubernetes.io/docs"}},
	// DevTools
	{Name: "vite", Category: "devtools", DocHints: []string{"https://vitejs.dev/guide"}},
	{Name: "webpack", Category: "devtools", DocHints: []string{"https://webpack.js.org/concepts"}},
	{Name: "esbuild", Category: "devtools", DocHints: []string{"https://esbuild.github.io/api"}},
	{Name: "turbo", Category: "devtools", DocHints: []string{"https://turbo.build/repo/docs"}},
	{Name: "turborepo", Category: "devtools", DocHints: []string{"https://turbo.build/repo/docs"}},
}

// RegistryByName returns a KnownLibrary for the given name, or nil if not found.
func RegistryByName(name string) *KnownLibrary {
	for i := range Registry {
		if Registry[i].Name == name {
			return &Registry[i]
		}
	}
	return nil
}

// CategoryForLibrary looks up the category for a library by exact match first,
// then by prefix match (e.g. "react" matches "react-query"). Returns "" if unknown.
func CategoryForLibrary(library string) string {
	normalized := strings.ToLower(sanitizeName(library))
	// Exact match
	for _, lib := range Registry {
		if strings.ToLower(sanitizeName(lib.Name)) == normalized {
			return lib.Category
		}
	}
	// Prefix match: registry name is a prefix of the input (e.g. "react" → "react-query")
	for _, lib := range Registry {
		registryName := strings.ToLower(sanitizeName(lib.Name))
		if registryName != "" && strings.HasPrefix(normalized, registryName) {
			return lib.Category
		}
	}
	return ""
}
