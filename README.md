# Kjernekraft
Yoga studio management system in Oslo

## Architecture Overview

This application uses a scope-based file organization system with comprehensive multi-language support and modular template architecture.

### Template Organization

The template system is organized by scope of influence, making it easy to find and maintain related files:

```
handlers/templates/
├── layouts/
│   └── base.html          # Ein mal for alle sidor: hovud, <main>, botnlina.
│                          # Alt av stil ligg i static/css/kjernekraft.css.
├── pages/                 # Ei fil per sida: dashboard, admin, innlogging,
│                          # gloymt-passord, registrering, vilkaar, min-profil,
│                          # membership, klippekort, betaling, timeplan, arket
├── components/            # Attbrukande bitar paa tvers av sidor
│   ├── navigation/
│   │   └── navigation.html
│   └── common/            # dagmerke, week-grid, language-selector,
│                          # charges-/payment-methods-containerar, *-scripts
└── modules/               # Funksjonsbolkar etter forretningsomraade
    ├── dashboard/         # signed-up-classes, todays-classes,
    │                      # dashboard-membership, dashboard-klippekort, scripts
    ├── admin/             # admin-class-management (timereglane),
    │                      # admin-users-table (folk), admin-freeze-requests-table,
    │                      # admin-pricing-management, admin-membership-rules,
    │                      # admin-stats, admin-settings, admin-scripts
    └── membership/        # membership, charges, klippekort
```

### Component vs Module Distinction

- **Components** (`components/`): Reusable UI elements that can be used across multiple pages
  - Navigation, language selectors, the dagmerke day token
  - Think "building blocks" that many pages might need
  - All styling lives in `static/css/kjernekraft.css` — there are no
    per-component style templates (see `docs/DESIGN_GUIDELINES.md`)

- **Modules** (`modules/`): Feature-specific functionality grouped by business domain
  - Dashboard widgets, admin tools, membership features
  - Think "feature packages" that belong to specific areas of the app

### Language System

The application supports three languages with cookie-based persistence:

- **Norwegian Nynorsk** (nn) 🇳🇴 - Default language
- **Norwegian Bokmål** (nb) 🇩🇰
- **English** (en) 🇺🇸

Language preferences are:
1. Stored in cookies with 1-year expiration
2. Available on all pages including login (before authentication)
3. Automatically detected from cookies or URL parameters
4. Integrated into user profile for easy switching

### Adding New Features

1. **New Page**: Create in `pages/` directory using `{{define "content"}}` 
2. **Reusable Component**: Add to `components/common/` for cross-page use
3. **Feature Module**: Create in `modules/[feature-name]/` for domain-specific functionality
4. **Localization**: Add keys to all three language files in `locales/`

### Development Guidelines

- All pages use the base layout system - no custom HTML structure
- All text must use localization keys - no hardcoded strings
- Related files (HTML, CSS, JS) should be co-located by feature
- Use the scope hierarchy to determine file placement
    │   ├── membership.css
    │   ├── klippekort.html
    │   ├── klippekort.css
    │   ├── charges.html
    │   └── charges.css
    └── events/       # Event-related modules
        ├── event_card.html
        └── event_card.css
```

### Organizational Principles

**Scope of Influence Hierarchy:**
1. **Layouts** - Site-wide impact (affects all pages)
2. **Pages** - Page-specific impact (single page)
3. **Components** - Cross-page impact (used across multiple pages)
4. **Modules** - Feature-specific impact (related to specific functionality)

**File Grouping Rules:**
- Files that do similar things are placed together
- Related CSS and HTML files are co-located in the same directory
- Module directories group all files related to a specific feature
- No duplication - each template has a single authoritative location

**Adding New Features:**
1. Page templates go in `pages/`
2. Cross-page UI elements go in `components/`
3. Feature-specific modules go in `modules/{feature_name}/`
4. Keep related CSS files in the same directory as their HTML templates

### Handler Organization

```
handlers/
├── config/           # Configuration management
├── modules/          # Go modules for template data
├── template_manager.go # Central template loading and management
├── localization.go   # Multi-language support
├── dashboard.go      # Dashboard page handlers
├── admin.go         # Admin panel handlers
├── membership.go    # Membership, klippekort, and profile handlers
├── users.go         # User authentication handlers
├── timeplan.go      # Schedule page handlers
├── betaling.go      # Payment page handlers
└── ...              # Other feature-specific handlers
```

### Localization System

The application supports three languages:
- **Norwegian Nynorsk** (`nn`) - Default
- **Norwegian Bokmål** (`nb`)
- **English** (`en`)

All user-facing text uses localization keys:
```html
{{t .Lang "navigation.home"}}
{{t .Lang "admin.approve"}}
```

Translation files are located in `locales/` directory. Do not take this
section's word for the state of them — `go test ./handlers` does. It fails if
the three files drift apart on keys, if a translation is empty, or if a
template asks for a key nobody has translated (which renders as the bare key,
silently).

## Development Guidelines

### Template Development
- Use the modular system - compose pages from existing modules when possible
- Add the `Lang` field to all handler data structures for localization
- Keep templates small and focused on a single responsibility
- Use descriptive template names that match their function

### Adding New Pages
1. Create page template in `pages/`
2. Create handler function that includes `Lang` field
3. Compose page from existing modules or create new ones as needed
4. Add localization keys for any new text

### File Organization
- Respect the scope hierarchy when placing files
- Keep related files together (HTML, CSS, and any assets)
- Avoid creating deep nested directories
- Use clear, descriptive names for directories and files

## Getting Started

The server reads two environment variables and refuses to start without the
first one:

```bash
export KJERNEKRAFT_SESSION_KEY=$(openssl rand -base64 32)
export KJERNEKRAFT_ENV=development

go run .
# http://localhost:8080
```

See `.env.example`. Generate a **different** key for production, keep it out of
the repository, and leave `KJERNEKRAFT_ENV` unset there — anything other than
`development` is treated as production, which turns on `Secure` cookies and
makes the test-data routes return 404.

The application starts in Norwegian Nynorsk — that is what you get when you
have not chosen. Add `?lang=en` or `?lang=nb` to any URL to switch languages;
the choice is remembered in a cookie.

## Access control

Authorization lives in the router, not in the handlers. `server.go` declares
three groups, and a route's placement is what grants it:

| Group | Middleware | Contains |
|---|---|---|
| open | — | `/`, `/signup`, `/terms`, `/innlogging`, `/logout`, `/static/*` |
| signed in | `RequireAuth` | `/elev/*` and the `/api/*` endpoints a student uses |
| admin | `RequireAdmin` | `/admin`, `/api/admin/*`, `/users/assign-role` |
| development | `RequireDevelopment` | the test-data routes, 404 outside `KJERNEKRAFT_ENV=development` |

Two rules hold this together:

- **Roles come from the database, once per request.** The session cookie holds
  a user ID and nothing else. `WithUser` loads the user, and `RequireAdmin`
  reads the role off that. A role in the cookie is a role the browser can
  rewrite.
- **Every mutating request carries a CSRF token.** `CSRF` issues it into the
  session and mirrors it into a readable cookie; `static/js/csrf.js` attaches
  it to htmx and `fetch` calls, and HTML forms carry it in a hidden
  `csrf_token` field. A `POST` without it gets 403.

Adding a route outside a group makes it public. That is the one thing to get
right in this file.

## Utviklingstenaren

```sh
./køyr              # startar, og byggjer paa nytt naar Go-filer endra seg
./køyr --ein-gong   # startar utan vaktar
./køyr --stogg      # stoggar
```

Han lyder paa <http://localhost:8080> og skriv til `.køyr.logg`.

**Kva som krev kva:**

| Du endrar | Kva som skjer |
|---|---|
| `.go` | vaktaren byggjer og startar paa nytt, kring eitt sekund |
| malar i `handlers/templates/` | ingen ting — dei vert lesne for kvar soknad. Oppdater sida. |
| `static/css`, `static/js` | ingen ting. Oppdater sida. |

Dei tvo siste gjeld berre naar `KJERNEKRAFT_ENV=development`, som `./køyr`
set. Statiske filer vert sende med `Cache-Control: no-store` i utvikling —
elles bufrar lesaren dei, og ein sit og ser paa si eigi gamle CSS medan
ein lurer paa kvifor endringi ikkje slo inn.

Skriptet set ogso ein fast `KJERNEKRAFT_SESSION_KEY`, so ein ikkje vert
logga ut kvar gong tenaren startar. Han er **berre** til dette; i drift
kjem nykelen or umgjevnaden og skal vera tilfeldig.

### Prøvebrukar

`anna@example.com` / `password123` — ho hev admin-rolla, so
<http://localhost:8080/admin> er open for henne.
