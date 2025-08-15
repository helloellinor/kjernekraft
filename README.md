# Kjernekraft
Yoga studio management system in Oslo

## Architecture Overview

This application uses a scope-based file organization system with comprehensive multi-language support and modular template architecture.

### Template Organization

The template system is organized by scope of influence, making it easy to find and maintain related files:

```
handlers/templates/
├── layouts/            # Site-wide scope (affects all pages)
│   └── base.html      # Main layout with navigation, styles, and structure
├── pages/             # Page-specific scope (individual page templates)
│   ├── dashboard.html # User dashboard
│   ├── admin.html     # Admin dashboard  
│   ├── innlogging.html # Login page
│   ├── min-profil.html # User profile
│   ├── membership.html # Membership selection
│   ├── klippekort.html # Punch cards
│   ├── betaling.html  # Payments
│   └── timeplan.html  # Schedule
├── components/        # Cross-page scope (reusable UI elements)
│   ├── navigation/    # Site navigation
│   │   ├── navigation.html
│   │   └── navigation-styles.html
│   └── common/        # Shared components and styles
│       ├── common-styles.html      # Base styles for all pages
│       ├── event-card-styles.html  # Event card styling
│       ├── language-selector.html  # Language switching component
│       ├── profile-styles.html     # Profile page styles
│       ├── betaling-styles.html    # Payment page styles
│       └── login-styles.html       # Login page styles
└── modules/           # Feature-specific scope (related functionality)
    ├── dashboard/     # Dashboard-specific modules
    │   ├── signed-up-classes.html
    │   ├── todays-classes.html
    │   ├── dashboard-membership.html
    │   ├── dashboard-klippekort.html
    │   └── dashboard-scripts.html
    ├── admin/         # Admin panel modules
    │   ├── admin-users-table.html
    │   ├── admin-events-table.html
    │   ├── admin-freeze-requests-table.html
    │   ├── admin-scripts.html
    │   └── admin-styles.html
    ├── membership/    # Membership management
    │   ├── membership.html
    │   ├── charges.html
    │   └── klippekort.html
    └── events/        # Event-related functionality
        └── event_card.html
```

### Component vs Module Distinction

- **Components** (`components/`): Reusable UI elements that can be used across multiple pages
  - Navigation, styles, language selectors, common layouts
  - Think "building blocks" that many pages might need

- **Modules** (`modules/`): Feature-specific functionality grouped by business domain
  - Dashboard widgets, admin tools, membership features
  - Think "feature packages" that belong to specific areas of the app

### Language System

The application supports three languages with cookie-based persistence:

- **Norwegian Bokmål** (nb) 🇩🇰 - Default language
- **Norwegian Nynorsk** (nn) 🇳🇴  
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
- **Norwegian Bokmål** (`nb`) - Default
- **Norwegian Nynorsk** (`nn`)
- **English** (`en`)

All user-facing text uses localization keys:
```html
{{t .Lang "navigation.home"}}
{{t .Lang "admin.approve"}}
```

Translation files are located in `locales/` directory.

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

```bash
# Start the application
go run server.go

# Access the application
http://localhost:8080
```

The application will start with default Norwegian Bokmål language. Add `?lang=en` or `?lang=nn` to any URL to switch languages.
