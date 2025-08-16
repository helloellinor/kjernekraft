# Design Guidelines for Kjernekraft

## Visual Design Principles

### Shadows Over Borders

The Kjernekraft design system uses shadows instead of solid borders for modern, elevated UI components. This creates a softer, more sophisticated appearance while maintaining clear visual hierarchy.

**Implementation:**
- **Primary Shadows**: `box-shadow: 0 4px 12px rgba(0,0,0,0.1)` for main containers
- **Hover Shadows**: `box-shadow: 0 4px 15px rgba(0,0,0,0.15)` for interactive elements
- **State-based Colored Shadows**: Use colored shadows to indicate different states
  - Active membership: `box-shadow: 0 4px 12px rgba(40, 167, 69, 0.3)`
  - Paused membership: `box-shadow: 0 4px 12px rgba(0, 124, 186, 0.3)`
  - Expiring klippekort: `box-shadow: 0 4px 12px rgba(255, 107, 53, 0.3)`

**Components Using Shadow System:**
- Membership cards
- Klippekort cards  
- Event cards
- Payment method cards
- Admin panels
- Form elements (select dropdowns)

### Badge Positioning

Status badges and special offer indicators follow consistent positioning:
- **Top-left**: Status badges (e.g., "AKTIV", "MEST POPULÆR")
- **Top-right**: Special offer badges (e.g., "Spesialtilbud")
- **Under title**: Important metadata (e.g., binding period information)

### Mobile Responsiveness

All dashboard modules automatically convert to single-column layout on mobile devices:
```css
@media (max-width: 768px) {
    .selector-container {
        grid-template-columns: 1fr;
        gap: 2rem;
    }
}
```

### Color Scheme

**Primary Colors:**
- Main blue: `#007cba`
- Success green: `#28a745`
- Warning orange: `#ff6b35`
- Error red: `#dc3545`

**Shadow Colors:**
- Default: `rgba(0,0,0,0.1)`
- Active: `rgba(40, 167, 69, 0.3)`
- Warning: `rgba(255, 107, 53, 0.3)`
- Info: `rgba(0, 124, 186, 0.3)`

## Implementation Notes

When creating new components, always use shadows instead of borders unless specifically needed for functional purposes (like form field focus states or color-coded left borders on event cards).

All UI text should use localization keys from the centralized translation system to support multiple languages (Norwegian Bokmål, Nynorsk, and English).