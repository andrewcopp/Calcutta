# UX Reviewer Memory - Calcutta Frontend

## Project Context
- Data science/research tool for technically savvy users
- Lab page tracks ML investment models through pipeline: Model → Entry → Evaluation
- Each model is tested against ~8 historical calcuttas as reference points
- Users prioritize data density over visual polish

## Component Patterns
- **PageContainer**: `py-8` padding (32px vertical)
- **PageHeader**: `mb-8` margin (32px bottom), includes title (text-3xl) + subtitle
- **Card**: `p-6 shadow rounded-lg` (24px padding)
- **TabsNav**: `mb-8` margin, lives inside its own Card wrapper
- Tab implementation has proper ARIA and keyboard navigation (arrow keys, Home/End)

## Lab Page Architecture
- 3 tabs: "By Model" (default), "By Calcutta", "Leaderboard"
- "By Model" shows expandable cards with pipeline matrix
- Each model card shows: name, kind, avg metrics, completion status
- Expanded view shows table: Calcutta | Entry | Evaluation columns
- PipelineStatusCell uses icons: ✓ (complete), ⏳ (pending), − (missing)

## Known Issues Addressed
- Previous iteration wasted excessive vertical space on header/tabs
- Conceptual model was inverted (organized by Calcutta instead of by Model)
- Fixed: reorganized around pipeline stages with calcuttas as data points
