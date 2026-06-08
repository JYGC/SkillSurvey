# Requirements: Auth Gate, Sidebar Move, and Bootstrap-to-Carbon Migration

## Auth Gate for MonthlyCountReport

### Route protection
WHEN an unauthenticated user navigates to `/monthly-count-report` THE SYSTEM SHALL redirect them to the login page.
WHEN an authenticated user navigates to `/user/monthly-count-report` THE SYSTEM SHALL display the monthly count report.
WHEN an authenticated user logs in successfully THE SYSTEM SHALL redirect them to `/user/monthly-count-report`.

## Sidebar Navigation

### Sidebar location
WHEN an authenticated user views any page under `/user` THE SYSTEM SHALL display a persistent sidebar with navigation links.
WHEN an unauthenticated user views the login page THE SYSTEM SHALL NOT display a sidebar.
WHEN an unauthenticated user views the register page THE SYSTEM SHALL NOT display a sidebar.

### Sidebar contents (authenticated)
WHEN the sidebar is displayed THE SYSTEM SHALL include a link to "Monthly count report" navigating to `/user/monthly-count-report`.
WHEN the sidebar is displayed THE SYSTEM SHALL include a link to "Settings" navigating to `/user/settings`.
WHEN the sidebar is displayed THE SYSTEM SHALL highlight the link matching the current route as active.

## Carbon Design System

### Exclusive Carbon usage
WHEN any UI element is rendered THE SYSTEM SHALL use Carbon Design System components (`@carbon/vue`) exclusively.
WHEN any layout, form, button, input, link, or navigation element is rendered THE SYSTEM SHALL NOT use Bootstrap or Bootstrap Vue 3 components or CSS classes.

### Header (authenticated)
WHEN an authenticated user views any page under `/user` THE SYSTEM SHALL display a Carbon header bar showing the logged-in user's email and a Logout button.
WHEN the user clicks Logout THE SYSTEM SHALL log them out and redirect to `/`.

## PocketBase API Rules

### monthlyCountReports collection
WHEN an unauthenticated request is made to list or view `monthlyCountReports` records THE SYSTEM SHALL deny access (HTTP 403).
WHEN an authenticated request is made to list or view `monthlyCountReports` records THE SYSTEM SHALL allow access.
WHEN `runtask report` writes `monthlyCountReports` records THE SYSTEM SHALL continue to allow write access (write rules are unchanged).

## Unchanged Behaviours

WHEN an authenticated user visits a public route (`/`, `/login`, `/register`) THE SYSTEM SHALL redirect to `/user/monthly-count-report`.
WHEN an unauthenticated user visits any user route (`/user/*`) THE SYSTEM SHALL redirect to `/`.
WHEN an unauthenticated user visits `/login` THE SYSTEM SHALL display the login form.
WHEN an unauthenticated user visits `/register` THE SYSTEM SHALL display the registration form.
