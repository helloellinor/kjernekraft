# Kjernekraft Template System Style Guide

This document defines the coding standards and best practices for the Kjernekraft template system to ensure consistency, maintainability, and high code quality across the entire project.

## 📋 File Naming Conventions

### HTML Templates
- Use **kebab-case** for all HTML template files
- Be descriptive and specific about the component's purpose
- Include the component type when helpful

```
✅ Good:
- user-profile-card.html
- admin-settings-form.html
- event-card.html

❌ Bad:
- userProfileCard.html
- admin_settings.html
- eventcard.html
```

### CSS Files
- Use **kebab-case** for CSS files
- Match the naming of corresponding HTML templates
- Include `-styles` suffix for clarity

```
✅ Good:
- button-styles.css
- event-card-styles.css
- admin-form-styles.css

❌ Bad:
- buttonStyles.css
- event_card.css
- AdminForm.css
```

### Directory Names
- Use **kebab-case** for directory names
- Use descriptive, domain-specific names
- Group related functionality together

## 🏗️ Component Organization

### Component Types

#### 1. UI Components (`/components/ui/`)
Generic, reusable interface elements with no business logic.

**Characteristics:**
- Highly reusable across different contexts
- Accept configuration through template variables
- Minimal to no business logic
- Focus on presentation and interaction

**Examples:**
- `button.html` - Generic button component
- `card.html` - Container component
- `form.html` - Form layout component

#### 2. Feature Components (`/components/features/`)
Domain-specific components containing business logic.

**Organization by Domain:**
- `auth/` - Authentication and user management
- `membership/` - Membership and subscription logic
- `admin/` - Administrative interfaces
- `dashboard/` - User dashboard functionality
- `klippekort/` - Punch card system
- `events/` - Event and class management

#### 3. Layout Components (`/components/layout/`)
Structural components for page organization.

#### 4. Navigation Components (`/components/navigation/`)
Navigation, menus, and language selection.

### Component Structure Guidelines

Each component should:
1. Have a single, well-defined responsibility
2. Be named clearly to indicate its purpose
3. Include corresponding styles when needed
4. Be documented with clear usage examples

## 🎨 Styling Architecture

### Style Organization Hierarchy

```
core/styles/
├── shared-styles.html          # Global utilities and base styles
├── common-styles.html          # Common design patterns
├── {feature}-styles.html       # Page-specific styles
└── components/                 # Component-specific styles
    ├── ui/                     # UI component styles
    └── features/               # Feature component styles
        ├── membership/
        ├── admin/
        ├── events/
        └── ...
```

### Style Guidelines

1. **Separation**: Keep styles separate from templates
2. **Specificity**: Use component-specific styles when possible
3. **Reusability**: Leverage shared styles for common patterns
4. **Naming**: Follow BEM or similar methodology for CSS classes

## 📝 Template Coding Standards

### Template Definition
```html
{{define "component-name"}}
<!-- Component content -->
{{end}}
```

### Variable Usage
- Use descriptive variable names
- Document expected variables in component comments
- Provide sensible defaults when possible

### Comments
```html
{{/* Component: User Profile Card
     Purpose: Display user information in a card format
     Variables: .User (required), .ShowActions (optional) */}}
{{define "user-profile-card"}}
<!-- Component implementation -->
{{end}}
```

### Error Handling
- Components should gracefully handle missing data
- Use conditional blocks to prevent template errors
- Provide fallback content when appropriate

## 🔧 Development Workflow

### Adding New Components

1. **Planning Phase**
   - Determine component type (UI vs Feature)
   - Choose appropriate directory location
   - Plan component interface (variables needed)

2. **Implementation Phase**
   - Create HTML template following naming conventions
   - Add corresponding styles if needed
   - Include documentation comments

3. **Integration Phase**
   - Update relevant page templates
   - Test component in different contexts
   - Update documentation if needed

### Modifying Existing Components

1. **Impact Assessment**
   - Identify all files using the component
   - Plan backward compatibility strategy
   - Consider creating new version if breaking changes needed

2. **Implementation**
   - Make minimal, focused changes
   - Maintain existing interface when possible
   - Update documentation and comments

3. **Testing**
   - Test all pages using the component
   - Verify no template errors introduced
   - Check visual consistency

## 🧪 Quality Assurance

### Component Checklist

Before submitting a component:

- [ ] Follows naming conventions
- [ ] Located in appropriate directory
- [ ] Includes documentation comments
- [ ] Has corresponding styles (if needed)
- [ ] Tested in multiple contexts
- [ ] Handles edge cases gracefully
- [ ] No template parsing errors

### Code Review Guidelines

When reviewing template code:

1. **Structure**: Is the component well-organized and logical?
2. **Reusability**: Can this component be used in other contexts?
3. **Performance**: Are there any inefficient template operations?
4. **Maintainability**: Is the code easy to understand and modify?
5. **Standards**: Does it follow the established conventions?

## 📚 Documentation Requirements

### Component Documentation

Each complex component should include:

1. **Purpose**: What does this component do?
2. **Interface**: What variables does it expect?
3. **Usage Examples**: How should it be used?
4. **Dependencies**: What other components does it require?

### Example Documentation
```html
{{/* 
Component: Event Card
Purpose: Display event information in a card format with booking actions
Variables:
  - .Event (required): Event object with Title, Date, Instructor
  - .ShowBooking (optional): Whether to show booking button (default: true)
  - .CompactMode (optional): Use compact layout (default: false)

Usage:
  {{template "event-card" .}}
  {{template "event-card" (dict "Event" .Event "ShowBooking" false)}}

Dependencies:
  - button.html
  - card-styles.html
*/}}
```

## 🚀 Performance Guidelines

### Template Performance

1. **Minimize Template Calls**: Reduce nested template calls when possible
2. **Cache Template Results**: Use variables to store computed values
3. **Conditional Loading**: Only load components when needed
4. **Optimize Loops**: Be efficient in range operations

### Style Performance

1. **CSS Organization**: Group related styles together
2. **Minimize Duplication**: Use shared styles for common patterns
3. **Optimize Selectors**: Use efficient CSS selectors
4. **Minify in Production**: Ensure styles are optimized for production

## 🔄 Migration Guidelines

When upgrading or refactoring components:

1. **Backward Compatibility**: Maintain existing interfaces when possible
2. **Deprecation Warnings**: Add warnings for deprecated features
3. **Migration Path**: Provide clear upgrade instructions
4. **Testing**: Thoroughly test all affected pages

This style guide ensures our template system remains consistent, maintainable, and scalable as the project grows.