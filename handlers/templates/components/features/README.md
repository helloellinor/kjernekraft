# Feature Components

This directory contains business domain-specific components that implement the core functionality of the Kjernekraft application.

## Domain Organization

### Authentication (`auth/`)
Components for user authentication, login, and session management.

### Membership (`membership/`)
Components for managing user memberships, subscription plans, and billing.
- Membership cards and status displays
- Subscription selection interfaces
- Billing and charges management
- Klippekort (punch card) integration

### Administration (`admin/`)
Administrative interface components for system management.
- User management tables
- System statistics and analytics
- Configuration interfaces
- Administrative actions and forms

### Dashboard (`dashboard/`)
User dashboard components for displaying personalized information.
- Class schedules and upcoming events
- User's registered classes
- Membership status overview
- Quick action interfaces

### Klippekort (`klippekort/`)
Punch card system components for class-based billing.
- Punch card displays and management
- Purchase interfaces
- Usage tracking and history

### Events (`events/`)
Event and class management components.
- Event cards and listings
- Class scheduling interfaces
- Event details and registration

## Usage Guidelines

- Each domain directory is self-contained
- Components within a domain can reference each other
- Cross-domain dependencies should be minimal
- Place shared functionality in `/components/ui/` instead