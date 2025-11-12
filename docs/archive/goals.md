# Project Goals - Markdown Export Feature

## Primary Goal

Add LLM-friendly Markdown export capability to confluence-reader, enabling users to:
1. Export Confluence content as clean, readable Markdown
2. Use git to track changes over time (version-friendly format)
3. Feed Confluence documentation to LLMs for analysis and Q&A

## Success Criteria

### Must Have (v2.0)
- [x] Convert Confluence HTML storage format to Markdown
- [x] Preserve document structure (headings, lists, code blocks, tables)
- [x] Handle Confluence-specific macros gracefully
- [x] Generate YAML frontmatter with page metadata
- [x] Export both HTML and Markdown (or Markdown only, configurable)
- [x] CLI flag to enable markdown export (`--export-markdown` or env var)
- [x] Documentation for users

### Should Have
- [ ] Git-friendly output (consistent formatting, stable diffs)
- [ ] LLM-optimized format (clear structure, good context preservation)
- [ ] Internal link conversion (Confluence page links → relative MD links)
- [ ] Breadcrumbs from page hierarchy
- [ ] Attachment references with relative paths

### Nice to Have
- [ ] Incremental sync (only convert changed pages)
- [ ] Git auto-commit after sync
- [ ] Index/TOC generation from space structure
- [ ] Change detection (track page versions)

## Non-Goals (Out of Scope)

- Bidirectional sync (Markdown → Confluence)
- Real-time sync (webhooks)
- Web UI for browsing exported content
- Search functionality (use git grep or LLM)
- Custom Markdown flavors (stick to GitHub Flavored Markdown)

## Target Users

1. **Developers** - Want to grep/search Confluence docs locally
2. **AI/LLM users** - Feed documentation to ChatGPT, Claude, etc.
3. **Git enthusiasts** - Track documentation changes over time
4. **Backup users** - Human-readable backup format

## Measure of Success

- [x] Can export entire Confluence instance to Markdown
- [x] Markdown is clean (no HTML artifacts)
- [ ] LLMs can understand the content without confusion
- [ ] Git diffs show meaningful changes (not formatting noise)
- [ ] Users report satisfaction with output quality
- [ ] Performance: <1 minute for typical instance sync

---

**Status**: Phase 2 Complete (Integration)  
**Next**: Phase 3 (End-to-End Testing & Documentation)
