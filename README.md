# Kjernekraft
Yoga studio management system in Oslo

## File Structure Organization

The project follows a clean architecture with files organized by scope of influence. This makes it easy to find related files and understand the system's modular structure.

### Template Organization

```
handlers/templates/
├── layouts/           # Site-wide layouts
│   └── base.html     # Main page layout with navigation and common structure
├── pages/            # Complete page templates (page-specific scope)
│   ├── admin.html    # Admin dashboard page
│   ├── dashboard.html # User dashboard page
│   ├── innlogging.html # Login page
│   ├── membership.html # Membership selection page
│   ├── klippekort.html # Punch cards page
│   ├── betaling.html  # Payment page
│   ├── timeplan.html  # Schedule page
│   └── min-profil.html # User profile page
├── components/       # Reusable UI components (cross-page scope)
│   ├── navigation/   # Navigation component and styles
│   │   ├── navigation.html
│   │   └── navigation-styles.html
│   └── common/       # Common styles and components
│       ├── button-styles.html
│       ├── common-styles.html
│       ├── module-styles.html
│       └── standard-module.html
└── modules/          # Feature-specific modules (feature scope)
    ├── dashboard/    # Dashboard-specific modules
    │   ├── signed-up-classes.html
    │   ├── todays-classes.html
    │   ├── dashboard-membership.html
    │   ├── dashboard-klippekort.html
    │   ├── dashboard-scripts.html
    │   └── dashboard-layout.html
    ├── admin/        # Admin panel modules
    │   ├── admin-users-table.html
    │   ├── admin-events-table.html
    │   ├── admin-freeze-requests-table.html
    │   ├── admin-scripts.html
    │   ├── admin-styles.html
    │   ├── admin-stats.html
    │   ├── admin-stats.css
    │   └── admin_settings.html
    ├── membership/   # Membership and payment modules
    │   ├── membership.html
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
