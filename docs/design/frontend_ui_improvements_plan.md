# Frontend UI Improvements Plan

This document tracks follow-up UI/design improvements after the `PageContainer`/`PageHeader`/`Card` migration.

## Cross-cutting consistency
- [ ] Standardize loading state UI across pages (always `PageContainer` + `LoadingState`)
- [ ] Standardize error state UI across pages (always `Alert variant="error"` + retry when possible)
- [ ] Standardize empty state UI across pages (use `Alert variant="info"` or a consistent empty state pattern)
- [ ] Audit and remove remaining ad-hoc layout wrappers now that `PageContainer` is the default

## Tabs UX + Accessibility (`TabsNav`)
- [ ] Add `role="tablist"` / `role="tab"` semantics and `aria-selected`/roving `tabIndex`
- [ ] Add keyboard navigation (ArrowLeft/ArrowRight/Home/End) for tabs
- [ ] Add `focus-visible` styles for tab buttons
- [ ] (Optional) Add sticky tabs support for long pages (without breaking existing layouts)

## Tables
- [ ] Convert remaining raw `<table>` instances to `components/ui/Table` primitives where feasible
- [ ] Add consistent numeric alignment utilities (e.g. `text-right`) and table density guidelines
- [ ] (Optional) Add a lightweight sortable header pattern (visual affordance only)

## Forms
- [ ] Standardize label + help text + error text spacing (consider a `FormField` wrapper)
- [ ] Standardize disabled/busy button states for mutations
- [ ] Ensure validation messaging uses consistent `Alert`/inline styles

## Typography + Spacing
- [ ] Define a consistent heading scale for pages/cards (h1/h2/h3 usage)
- [ ] Normalize vertical rhythm (`space-y-*`, `gap-*`) across migrated pages
- [ ] Audit `Card` usage to reduce nested cards / inconsistent padding

## Accessibility & Motion
- [ ] Ensure icon-only buttons have `aria-label`
- [ ] Ensure interactive controls have visible focus rings (`focus-visible:*`)
- [ ] Respect reduced-motion preferences for animations/transitions across pages

## Execution workflow
- [ ] Make changes in small, focused frontend-only commits
- [ ] Run `npm run build` in `frontend/` before each commit
- [ ] Avoid staging or modifying backend/unrelated files while doing UI work
