# Site Localization Guide

This document describes the complete process for localizing the Kjernekraft website, including adding new languages, translating content, and maintaining the localization system.

## Overview

The Kjernekraft website supports three languages:
- **Norwegian Bokmål** (`nb`) - Default language
- **Norwegian Nynorsk** (`nn`) 
- **English** (`en`)

All user-facing text is stored in JSON translation files and accessed through template functions for consistent multilingual support.

## File Structure

```
locales/
├── nb/
│   └── common.json     # Norwegian Bokmål translations
├── nn/
│   └── common.json     # Norwegian Nynorsk translations
└── en/
    └── common.json     # English translations
```

## Translation File Format

Translation files use nested JSON structure for organization:

```json
{
  "section_name": {
    "key": "Translated text",
    "nested_section": {
      "sub_key": "More translated text"
    }
  }
}
```

### Example Structure

```json
{
  "navigation": {
    "home": "Hjem",
    "schedule": "Timeplan",
    "logout": "Logg ut"
  },
  "dashboard": {
    "welcome_user": "Velkommen, {{.UserName}}!",
    "loading_membership": "Laster medlemskap..."
  }
}
```

## Using Translations in Templates

### Basic Translation Function

Use the `t` function to access translations:

```html
<h1>{{t .Lang "navigation.home"}}</h1>
<p>{{t .Lang "dashboard.loading_membership"}}</p>
```

### Translation Function in JavaScript

For JavaScript strings, use the `toJS` function for proper escaping:

```html
<script>
const messages = {
    welcome: {{t .Lang "dashboard.welcome_user" | toJS}},
    error: {{t .Lang "errors.general" | toJS}}
};
</script>
```

### Available Template Functions

- `t` - Main translation function: `{{t .Lang "key.path"}}`
- `translate` - Alias for `t`: `{{translate .Lang "key.path"}}`
- `toJS` - JavaScript-safe translation: `{{t .Lang "key" | toJS}}`

## Language Switching

Languages can be switched via URL parameter:
- Default: `/page` (uses Norwegian Bokmål)
- English: `/page?lang=en`
- Nynorsk: `/page?lang=nn`

## Adding a New Language

### 1. Create Language Directory

```bash
mkdir locales/[language_code]
```

### 2. Create Translation File

```bash
cp locales/nb/common.json locales/[language_code]/common.json
```

### 3. Update Localization System

In `handlers/localization.go`, add the new language to the supported languages list:

```go
languages := []string{"nb", "nn", "en", "new_language_code"}
```

### 4. Translate Content

Edit the new `common.json` file and translate all values while keeping the same key structure.

## Translation Key Organization

### Current Key Categories

#### Site-Wide Elements
```json
"site": {
  "title": "Site title",
  "welcome": "Welcome message"
}
```

#### Navigation
```json
"navigation": {
  "dashboard_title": "Page header title",
  "logout": "Logout button",
  "home": "Home link",
  "schedule": "Schedule link"
}
```

#### Dashboard Interface
```json
"dashboard": {
  "welcome_user": "Welcome message with user name",
  "signed_up": "Signed up section title",
  "loading_membership": "Loading state message"
}
```

#### User Actions & Confirmations
```json
"membership_actions": {
  "freeze_confirm": "Confirmation dialog text",
  "error_prefix": "Error message prefix"
}
```

#### Admin Interface
```json
"admin": {
  "title": "Admin page title",
  "user_table": {
    "name": "Table column header",
    "email": "Table column header"
  },
  "alerts": {
    "freeze_approved": "Success message"
  }
}
```

## Best Practices

### Key Naming Conventions

1. **Use descriptive, hierarchical keys**: `dashboard.membership.loading` not `loading1`
2. **Group related content**: Put all table headers under `table_name.headers`
3. **Separate by UI section**: `navigation.*`, `dashboard.*`, `admin.*`
4. **Include context in key names**: `button_save` not just `save`

### Translation Guidelines

1. **Keep key structure identical** across all language files
2. **Use placeholder syntax** for dynamic content: `"Welcome, {{.UserName}}!"`
3. **Maintain consistent tone** within each language
4. **Test all languages** after adding new keys

### Adding New Keys

When adding new translatable text:

1. **Choose descriptive key path**: `section.component.element`
2. **Add to ALL language files** simultaneously
3. **Use template function** in HTML: `{{t .Lang "new.key"}}`
4. **Test in all supported languages**

## Module Integration

### Using Localization in Modules

Each module template should receive language context:

```html
{{define "my_module"}}
<div class="module">
    <h2>{{t .Lang "module.title"}}</h2>
    <p>{{t .Lang "module.description"}}</p>
</div>
{{end}}
```

### JavaScript Module Localization

For modules with JavaScript, create localized text objects:

```html
{{define "module_scripts"}}
<script>
const MODULE_TEXTS = {
    confirmDelete: {{t .Lang "module.confirm_delete" | toJS}},
    successMessage: {{t .Lang "module.success" | toJS}},
    errorMessage: {{t .Lang "module.error" | toJS}}
};

function deleteItem() {
    if (confirm(MODULE_TEXTS.confirmDelete)) {
        // Delete logic
        alert(MODULE_TEXTS.successMessage);
    }
}
</script>
{{end}}
```

## Testing Localization

### Manual Testing

1. **Test all language variants**:
   - Visit `/page?lang=nb` (default)
   - Visit `/page?lang=nn` 
   - Visit `/page?lang=en`

2. **Check for missing translations**:
   - Look for displayed key names instead of translated text
   - Check browser console for translation errors

3. **Verify special characters** render correctly in all languages

### Automated Checks

Create a simple script to verify key consistency:

```bash
# Check that all language files have the same keys
diff <(jq -r 'paths | join(".")' locales/nb/common.json | sort) \
     <(jq -r 'paths | join(".")' locales/en/common.json | sort)
```

## Maintenance

### Regular Tasks

1. **Review translation completeness** when adding new features
2. **Update translations** when UI text changes
3. **Test language switching** functionality
4. **Backup translation files** before major changes

### Common Issues

- **Missing translations**: Add key to all language files
- **Broken placeholders**: Ensure `{{.Variable}}` syntax is correct
- **JavaScript errors**: Use `toJS` function for JS strings
- **Key conflicts**: Use unique, descriptive key paths

## Troubleshooting

### Translation Not Appearing

1. Check key exists in all language files
2. Verify key path spelling in template
3. Ensure `.Lang` variable is passed to template
4. Check for JSON syntax errors in translation files

### JavaScript Localization Issues

1. Use `toJS` function: `{{t .Lang "key" | toJS}}`
2. Check for quote escaping in translated strings
3. Verify JavaScript object syntax is valid

### Performance Considerations

- Translation files are loaded once at startup
- Large translation files may impact startup time
- Consider splitting into multiple files if needed

## Future Enhancements

Potential improvements to the localization system:

1. **Date/time localization** with proper formatting
2. **Number formatting** for different locales  
3. **Pluralization support** for count-dependent strings
4. **Translation validation** tools
5. **Hot-reloading** of translation files in development

---

This documentation should be updated whenever changes are made to the localization system or when new translation patterns are established.