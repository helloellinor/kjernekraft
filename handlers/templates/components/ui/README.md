# UI Components

This directory contains the base UI component library for the Kjernekraft application. These are generic, reusable interface elements that can be used across different features and pages.

## Component Categories

### Form Components
- `form.html` / `form-styles.html` - Base form layout and styling
- `button.html` / `button-styles.html` - Button component with various styles

### Layout Components
- `card.html` / `card-styles.html` - Container component for content grouping
- `week-controls.html` - Calendar navigation controls
- `week-grid.html` - Calendar grid layout component

### Container Components
- `charges-container.html` - Container for billing information
- `payment-methods-container.html` - Payment method selection container
- `standard-module.html` - Generic module container

## Design Principles

1. **Generic**: Components should not contain business logic
2. **Configurable**: Accept parameters to customize behavior and appearance
3. **Reusable**: Can be used across different features and pages
4. **Consistent**: Follow the established design system patterns

## Usage

UI components are meant to be the building blocks for more complex feature components. They should be generic enough to be useful in multiple contexts while providing a consistent look and feel across the application.