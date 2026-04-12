package skill

// KnownLibrary describes a library in the markdocs registry.
type KnownLibrary struct {
	Name     string
	Category string
	DocHints []string // suggested doc URL prefixes for hinting
}

// Registry is the built-in list of commonly-documented libraries.
// Used by `markdocs scan` to cross-reference detected dependencies.
var Registry = []KnownLibrary{
	// Frontend
	{Name: "react", Category: "frontend", DocHints: []string{"https://react.dev"}},
	{Name: "vue", Category: "frontend", DocHints: []string{"https://vuejs.org/guide"}},
	{Name: "next", Category: "frontend", DocHints: []string{"https://nextjs.org/docs"}},
	{Name: "nuxt", Category: "frontend", DocHints: []string{"https://nuxt.com/docs"}},
	{Name: "svelte", Category: "frontend", DocHints: []string{"https://svelte.dev/docs"}},
	{Name: "astro", Category: "frontend", DocHints: []string{"https://docs.astro.build"}},
	{Name: "remix", Category: "frontend", DocHints: []string{"https://remix.run/docs"}},
	{Name: "@angular/core", Category: "frontend", DocHints: []string{"https://angular.io/docs"}},
	// UI Libraries
	{Name: "shadcn-ui", Category: "ui", DocHints: []string{"https://ui.shadcn.com/docs"}},
	{Name: "@radix-ui/react-dialog", Category: "ui", DocHints: []string{"https://www.radix-ui.com/docs"}},
	{Name: "tailwindcss", Category: "ui", DocHints: []string{"https://tailwindcss.com/docs"}},
	{Name: "@mui/material", Category: "ui", DocHints: []string{"https://mui.com/material-ui"}},
	{Name: "chakra-ui", Category: "ui", DocHints: []string{"https://chakra-ui.com/docs"}},
	// Backend
	{Name: "express", Category: "backend", DocHints: []string{"https://expressjs.com"}},
	{Name: "fastify", Category: "backend", DocHints: []string{"https://fastify.dev/docs"}},
	// Database
	{Name: "prisma", Category: "database", DocHints: []string{"https://www.prisma.io/docs"}},
	{Name: "drizzle-orm", Category: "database", DocHints: []string{"https://orm.drizzle.team/docs"}},
	{Name: "@supabase/supabase-js", Category: "database", DocHints: []string{"https://supabase.com/docs"}},
	// Testing
	{Name: "vitest", Category: "testing", DocHints: []string{"https://vitest.dev/guide"}},
	{Name: "jest", Category: "testing", DocHints: []string{"https://jestjs.io/docs"}},
	{Name: "playwright", Category: "testing", DocHints: []string{"https://playwright.dev/docs"}},
	{Name: "cypress", Category: "testing", DocHints: []string{"https://docs.cypress.io"}},
	// Auth
	{Name: "next-auth", Category: "auth", DocHints: []string{"https://next-auth.js.org/getting-started"}},
	{Name: "@auth/core", Category: "auth", DocHints: []string{"https://authjs.dev/getting-started"}},
	// Payments
	{Name: "stripe", Category: "payments", DocHints: []string{"https://stripe.com/docs"}},
	// Infra
	{Name: "docker", Category: "infra", DocHints: []string{"https://docs.docker.com"}},
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
