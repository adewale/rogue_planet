# Rogue Planet - Launch TODO

This document tracks the tasks required for a successful open source launch.

**Target Launch Date:** TBD
**GitHub Repository:** https://github.com/adewale/rogue_planet
**Current Status:** Pre-Launch Preparation

---

## ✅ Already Complete

- [x] Core Functionality - Fully working feed aggregator
- [x] Documentation - README, QUICKSTART, WORKFLOWS, CONTRIBUTING
- [x] Testing - >75% coverage on core packages
- [x] Clean Codebase - No development artifacts, consistent structure
- [x] Examples - Working config and feeds examples
- [x] License - MIT license included

---

## 🔴 Critical (Blocking Launch)

### 1. GitHub Repository Setup
- [ ] Create public repository at `github.com/adewale/rogue_planet`
- [ ] Set repository description: "A modern feed aggregator written in Go, inspired by Venus/Planet"
- [ ] Add repository topics: `go`, `golang`, `rss`, `atom`, `feed-aggregator`, `planet`, `static-site-generator`
- [ ] Push initial code to main branch

### 2. Fix Module Path
- [ ] Update `go.mod`: Change `github.com/roguep/rogue_planet` to `github.com/adewale/rogue_planet`
- [ ] Update all import statements in Go files to use new module path
- [ ] Run `go mod tidy` to verify
- [ ] Test build: `go build ./cmd/rp`
- [ ] Test tests: `go test ./...`

### 3. Version Tag
- [ ] Ensure all changes are committed
- [ ] Create annotated tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
- [ ] Push tag to GitHub: `git push origin v1.0.0`

### 4. GitHub Release
- [ ] Create v1.0.0 release on GitHub
- [ ] Add release title: "Rogue Planet v1.0.0 - Initial Release"
- [ ] Copy content from CHANGELOG.md as release notes
- [ ] Mark as "Latest Release"
- [ ] Optional: Attach prebuilt binaries (macOS, Linux, Windows)

---

## 🟡 Important (Should Have Before Launch)

### 5. README Enhancements
- [ ] Add badges at top:
  - Go version: `![Go Version](https://img.shields.io/github/go-mod/go-version/adewale/rogue_planet)`
  - License: `![License](https://img.shields.io/github/license/adewale/rogue_planet)`
  - Go Report Card: `![Go Report Card](https://goreportcard.com/badge/github.com/adewale/rogue_planet)`
- [ ] Add screenshot or demo GIF of generated HTML output
- [ ] Verify all links work with new GitHub path

### 6. Installation Testing
- [ ] Test on clean machine (or container): `go install github.com/adewale/rogue_planet/cmd/rp@latest`
- [ ] Verify binary is installed: `rp version`
- [ ] Test basic workflow: init → add-feed → update
- [ ] Document any issues found

### 7. Community Files
- [ ] Create `.github/ISSUE_TEMPLATE/bug_report.md`
- [ ] Create `.github/ISSUE_TEMPLATE/feature_request.md`
- [ ] Create `.github/PULL_REQUEST_TEMPLATE.md`
- [ ] Optional: Add `CODE_OF_CONDUCT.md` (Contributor Covenant recommended)

### 8. Announcement Plan
- [ ] Draft announcement post (200-300 words):
  - What it is
  - Why it exists (problems it solves)
  - Key features
  - How to install
  - Link to repo
- [ ] Identify launch venues:
  - [ ] Hacker News (Show HN: format)
  - [ ] Reddit: r/golang
  - [ ] Reddit: r/selfhosted
  - [ ] Lobsters (if you have invite)
  - [ ] Go Forum: https://forum.golangbridge.org/
- [ ] Optional: Submit to Go Weekly newsletter
- [ ] Optional: Tweet or Mastodon post

---

## 🟢 Nice to Have (Post-Launch)

### 9. CI/CD
- [ ] Create `.github/workflows/test.yml` - Run tests on push/PR
- [ ] Create `.github/workflows/release.yml` - Build binaries on tag
- [ ] Optional: Add coverage reporting (Codecov or Coveralls)
- [ ] Add build status badge to README

### 10. Additional Documentation
- [ ] Add architecture diagram to README or docs/
- [ ] Expand comparison table with Planet Venus
- [ ] Create migration guide from Planet Venus
- [ ] Add FAQ section

### 11. Live Demo
- [ ] Host example planet somewhere public
- [ ] Add "Live Demo" link to README
- [ ] Keep it updated with interesting feeds

### 12. Community Growth
- [ ] Set up GitHub Discussions (optional alternative to issues)
- [ ] Create SECURITY.md with vulnerability reporting process
- [ ] Monitor GitHub stars and track growth
- [ ] Respond to all issues/PRs within 48 hours

---

## Recommended Launch Sequence

### Phase 1: Pre-Launch (1-2 hours)
1. Fix module path (Critical #2)
2. Create GitHub repository (Critical #1)
3. Push code to GitHub
4. Create v1.0.0 tag and release (Critical #3, #4)
5. Test installation works (Important #6)
6. Add basic GitHub templates (Important #7)

### Phase 2: Launch Day
1. Post announcement to Hacker News (Show HN)
2. Post to r/golang on Reddit
3. Post to r/selfhosted on Reddit
4. Monitor comments and respond quickly
5. Fix any immediate installation issues
6. Track engagement and feedback

### Phase 3: Post-Launch (First Week)
1. Respond to all issues and questions
2. Merge any quick-win PRs
3. Add CI/CD if project gains traction
4. Consider adding badges and demo
5. Submit to Go Weekly if momentum is good
6. Write post-mortem/lessons learned

---

## Launch Metrics to Track

- [ ] GitHub stars (target: 50+ in first week)
- [ ] Issues opened (shows engagement)
- [ ] Successful installations reported
- [ ] HN/Reddit upvotes and comments
- [ ]go install download count (pkg.go.dev shows this after ~24hrs)

---

## Blockers and Risks

**Known Issues:**
- None currently identified

**Potential Risks:**
- Installation issues on different platforms (mitigate: test on Linux/macOS/Windows)
- Module path conflicts (mitigate: verify path is available)
- Security concerns raised (mitigate: reference security features in README)

---

## Notes

- All critical tasks must be completed before any public announcement
- Important tasks should be done before launch for best first impression
- Nice-to-have tasks can be done post-launch based on community response
- Keep this document updated as tasks are completed

---

**Last Updated:** 2025-10-06
**Status:** Ready for Phase 1 (Pre-Launch)
