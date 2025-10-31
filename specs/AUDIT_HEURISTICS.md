# Code Audit Heuristics

⚠️ **This document has been superseded by two new documents:**

## 1. Project-Specific Audit Report
**📄 [docs/ROGUE_PLANET_AUDIT.md](../docs/ROGUE_PLANET_AUDIT.md)**

Contains comprehensive audit results for this specific project including:
- All findings (security, code quality, tests)
- Current state of the codebase
- Metrics and statistics
- Specific bugs found and fixed
- Recommendations for Rogue Planet

## 2. Generic Go Auditing Guide
**📄 [docs/GO_AUDITING_HEURISTICS.md](../docs/GO_AUDITING_HEURISTICS.md)**

Reusable heuristics for auditing any Go project including:
- Security patterns (SQL injection, XSS, SSRF)
- Resource management patterns
- Concurrency best practices
- Error handling guidelines
- Test quality metrics
- Automated tools and commands
- Not specific to Rogue Planet

---

## Document History

**Version 1.0 (2025-10-19):** Initial heuristics based on first audit pass
**Version 2.0 (2025-10-20):** Updated with full audit findings and validated metrics
**Version 3.0 (2025-10-20):** Split into project-specific audit and generic heuristics

---

## Migration Notes

This document previously contained both:
1. Rogue Planet-specific findings → Now in `docs/ROGUE_PLANET_AUDIT.md`
2. General Go auditing patterns → Now in `docs/GO_AUDITING_HEURISTICS.md`

The split allows:
- **Rogue Planet team** to track their specific audit status
- **Other projects** to use the generic heuristics without Rogue Planet details
- **AI agents** to use reusable patterns on any Go codebase

---

## Quick Links

### For Rogue Planet Developers
- Current audit status: [ROGUE_PLANET_AUDIT.md](../docs/ROGUE_PLANET_AUDIT.md)
- Outstanding issues: See "Remaining Recommendations" section
- Metrics dashboard: See "Metrics Dashboard" section

### For Other Projects
- Auditing checklist: [GO_AUDITING_HEURISTICS.md](../docs/GO_AUDITING_HEURISTICS.md)
- Quick start: See "Quick Start Checklist" section
- Automated workflow: See "Audit Workflow Template" section

### For AI Agents
- Self-check guide: [GO_AUDITING_HEURISTICS.md](../docs/GO_AUDITING_HEURISTICS.md)
- Tips for AI: See "Tips for AI Coding Agents" section
- Patterns to avoid: See "LLM-specific anti-patterns"

---

**Please use the new documents instead of this one.**
