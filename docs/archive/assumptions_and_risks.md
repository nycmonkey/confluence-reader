# Assumptions and Risks - Markdown Export

## Assumptions

### Technical Assumptions

1. **Confluence API Stability**
   - **Assumption**: Confluence Cloud API v2 will remain stable
   - **Validation**: API is production-ready, widely used
   - **Risk**: Low - Atlassian maintains backward compatibility
   - **Mitigation**: Version pinning, monitoring API changes

2. **HTML Storage Format Consistency**
   - **Assumption**: Confluence storage format is consistent across pages
   - **Validation**: ✅ Analyzed 479 pages, patterns are consistent
   - **Risk**: Low - Confluence uses consistent schema
   - **Mitigation**: Graceful degradation for unexpected HTML

3. **html-to-markdown Library Quality**
   - **Assumption**: Library handles all HTML correctly
   - **Validation**: ✅ Tested with 479 real pages, no failures
   - **Risk**: Low - Library is mature and well-tested
   - **Mitigation**: Preprocessing handles Confluence-specific elements

4. **Performance Acceptable**
   - **Assumption**: ~2ms per page conversion is fast enough
   - **Validation**: ✅ Tested, network I/O is bottleneck, not conversion
   - **Risk**: Very Low - Could handle 10x slower and still be fine
   - **Mitigation**: Already concurrent page downloads (5 at once)

5. **Disk Space Availability**
   - **Assumption**: Users have enough disk space for dual formats (HTML + MD)
   - **Validation**: Markdown is ~50% size of HTML, storage is cheap
   - **Risk**: Very Low - Can make MD-only mode optional
   - **Mitigation**: Document storage requirements, add flags

### Functional Assumptions

6. **Users Want Both Formats**
   - **Assumption**: Users want HTML (source of truth) + Markdown (readable)
   - **Validation**: To be confirmed with users
   - **Risk**: Low - Can make configurable (HTML only, MD only, both)
   - **Mitigation**: Add flags: `--html-only`, `--markdown-only`

7. **LLM Optimization is Valuable**
   - **Assumption**: Clean Markdown improves LLM understanding
   - **Validation**: To be validated in Phase 4
   - **Risk**: Medium - LLMs might be fine with HTML
   - **Mitigation**: User testing, feedback collection

8. **Git Tracking is Desirable**
   - **Assumption**: Users want to track changes with git
   - **Validation**: Common use case, but not validated
   - **Risk**: Low - Optional feature, doesn't hurt if unused
   - **Mitigation**: Document git workflow, make optional

### User Assumptions

9. **Users Have Confluence Cloud Access**
   - **Assumption**: Target users have Confluence Cloud (not Server/DC)
   - **Validation**: Tool only supports Cloud API v2
   - **Risk**: Low - Cloud is the future, Server is deprecated
   - **Mitigation**: Document clearly, consider Server support later

10. **Users Can Set Environment Variables**
    - **Assumption**: Users can set `CONFLUENCE_EXPORT_MARKDOWN=true`
    - **Validation**: Standard practice, well-documented
    - **Risk**: Very Low - Also support interactive prompts
    - **Mitigation**: Document both methods (env vars + prompts)

---

## Risks

### High Priority Risks

#### Risk 1: Breaking Existing Functionality
- **Description**: Integration might break existing clone behavior
- **Impact**: High - Users lose existing features
- **Probability**: Medium
- **Mitigation**:
  - Keep HTML export as default
  - Make markdown optional (flag/env var)
  - Extensive regression testing
  - No changes to existing code paths unless necessary
- **Status**: ⚠️ Monitor carefully during Phase 2

#### Risk 2: Internal Link Resolution Complexity
- **Description**: Converting Confluence page links to MD links requires page index
- **Impact**: Medium - Broken links in exported markdown
- **Probability**: Medium
- **Mitigation**:
  - Build page ID → slug mapping during clone
  - Use page titles for slug generation
  - Handle slug collisions (append ID if needed)
  - Validate links after export
- **Status**: ⏳ Defer to Phase 3

#### Risk 3: Complex Table Conversion Issues
- **Description**: Merged cells, nested tables might not convert well
- **Impact**: Medium - Some tables unreadable in markdown
- **Probability**: Medium
- **Mitigation**:
  - Test with table-heavy pages
  - Graceful degradation (keep table, just less pretty)
  - Document known limitations
  - Consider custom table handling
- **Status**: ⏳ Test in Phase 4

### Medium Priority Risks

#### Risk 4: Page Hierarchy for Children Links
- **Description**: Children macro needs parent-child relationships
- **Impact**: Medium - Missing child page lists
- **Probability**: High
- **Mitigation**:
  - Leave as comment placeholder for now
  - Build hierarchy map in Phase 3
  - Generate child link list dynamically
  - Add to frontmatter for context
- **Status**: ✅ Acknowledged, deferred to Phase 3

#### Risk 5: Attachment Path Resolution
- **Description**: Image/attachment references might break if paths change
- **Impact**: Low - Images don't display in markdown
- **Probability**: Low
- **Mitigation**:
  - Use relative paths from current structure
  - Validate attachment existence
  - Document attachment organization
  - Consider phase 5 restructuring carefully
- **Status**: ⚠️ Monitor in Phase 2

#### Risk 6: Memory Usage on Large Instances
- **Description**: Very large instances (10,000+ pages) might use too much memory
- **Impact**: Low - Tool crashes or slows down
- **Probability**: Low
- **Mitigation**:
  - Streaming approach (already implemented)
  - Process pages one at a time (no buffering)
  - Monitor memory in performance tests
  - Add progress checkpoints
- **Status**: ✅ Architecture is streaming, should be fine

### Low Priority Risks

#### Risk 7: Confluence API Rate Limits
- **Description**: Excessive API calls might trigger rate limiting
- **Impact**: Low - Clone fails or slows down
- **Probability**: Very Low
- **Mitigation**:
  - Already uses batch fetching (100 items)
  - Concurrent downloads limited to 5
  - Could add exponential backoff for 429s
  - Document rate limits
- **Status**: ✅ Low risk, monitor in testing

#### Risk 8: Cross-Platform Path Issues
- **Description**: File paths might behave differently on Windows
- **Impact**: Low - Files not saved correctly on Windows
- **Probability**: Very Low
- **Mitigation**:
  - Use `filepath.Join()` (cross-platform)
  - Test on Windows before release
  - Sanitize filenames (already done)
  - Use forward slashes in links
- **Status**: ✅ Already using filepath package

#### Risk 9: YAML Frontmatter Escaping
- **Description**: Page titles with quotes might break YAML
- **Impact**: Low - Invalid frontmatter
- **Probability**: Medium
- **Mitigation**:
  - Escape quotes in title (`"` → `\"`)
  - Use YAML library for generation (safer)
  - Validate YAML after generation
  - Test with special characters
- **Status**: ⚠️ Implement escaping in Phase 2

#### Risk 10: Dependency Version Conflicts
- **Description**: html-to-markdown library might conflict with other deps
- **Impact**: Low - Build failures
- **Probability**: Very Low
- **Mitigation**:
  - Go modules handle deps well
  - Library has minimal deps (2 packages)
  - Both deps are stable (golang.org/x/net)
  - Pin versions in go.mod
- **Status**: ✅ No conflicts observed

---

## Risk Matrix

| Risk | Impact | Probability | Priority | Status |
|------|--------|-------------|----------|--------|
| Breaking existing functionality | High | Medium | **Critical** | ⚠️ Monitor |
| Internal link resolution | Medium | Medium | High | ⏳ Phase 3 |
| Complex table conversion | Medium | Medium | Medium | ⏳ Phase 4 |
| Page hierarchy | Medium | High | Medium | ✅ Deferred |
| Attachment paths | Low | Low | Low | ⚠️ Monitor |
| Memory usage | Low | Low | Low | ✅ OK |
| API rate limits | Low | Very Low | Low | ✅ OK |
| Cross-platform paths | Low | Very Low | Low | ✅ OK |
| YAML escaping | Low | Medium | Low | ⚠️ Phase 2 |
| Dependency conflicts | Low | Very Low | Low | ✅ OK |

---

## Open Questions

### Questions Requiring Decisions

1. **Q: Should we support Confluence Server/Data Center?**
   - **Current**: Only Cloud API v2
   - **Decision needed**: Phase 7 (post-MVP)
   - **Impact**: Medium - expands user base, more work

2. **Q: Should we validate internal links after export?**
   - **Current**: Convert blindly, assume links work
   - **Decision needed**: Phase 4
   - **Impact**: Low - nice to have, not critical

3. **Q: Should we support markdown-only export (no HTML)?**
   - **Current**: Export both HTML and MD
   - **Decision needed**: Phase 2 (could add flag easily)
   - **Impact**: Low - user preference

4. **Q: Should we auto-commit to git after sync?**
   - **Current**: Manual git workflow
   - **Decision needed**: Phase 5
   - **Impact**: Low - convenience feature

### Questions for User Validation

1. **Q: Do users want both HTML and Markdown, or just Markdown?**
   - **Plan**: Ask in Phase 7 (UAT)
   - **Impact on design**: Flag for format selection

2. **Q: Do users care about page hierarchy in export?**
   - **Plan**: Ask in Phase 7 (UAT)
   - **Impact on design**: Breadcrumbs, child links

3. **Q: What's the typical Confluence instance size?**
   - **Plan**: Collect data in Phase 6
   - **Impact on design**: Performance optimization priorities

---

## Mitigation Strategies

### For High-Priority Risks

1. **Breaking Changes Prevention**
   ```bash
   # Before releasing Phase 2:
   - Run full regression test suite
   - Test export WITHOUT markdown flag
   - Verify HTML files still created correctly
   - Verify attachments still downloaded
   - Check that all existing tests pass
   ```

2. **Link Resolution Testing**
   ```bash
   # Before releasing Phase 3:
   - Export space with many internal links
   - Validate all links resolve correctly
   - Test slug collision handling
   - Document known link limitations
   ```

3. **Table Conversion Testing**
   ```bash
   # Before releasing Phase 4:
   - Collect pages with complex tables
   - Test conversion of each type
   - Document which table features work
   - Add warnings for known issues
   ```

### Contingency Plans

#### Plan A: If Conversion Quality is Poor
- **Trigger**: User feedback that markdown is unreadable
- **Action**: 
  1. Collect problematic pages
  2. Analyze common issues
  3. Improve preprocessing rules
  4. Consider custom HTML→MD rules
  5. Worst case: Keep HTML as primary, MD as bonus

#### Plan B: If Performance is Unacceptable
- **Trigger**: Export takes >5 minutes for typical instance
- **Action**:
  1. Profile conversion code
  2. Optimize hot paths
  3. Increase concurrency (5 → 10+ pages)
  4. Add progress indicators
  5. Worst case: Make markdown conversion optional/background

#### Plan C: If Link Resolution Fails
- **Trigger**: Many broken links in exported markdown
- **Action**:
  1. Keep Confluence URLs as fallback
  2. Add validation script to report broken links
  3. Generate link map file for manual fixing
  4. Worst case: Link to HTML files instead of MD

---

## Risk Review Schedule

- **Phase 2**: Review risks 1, 5, 9 (breaking changes, paths, YAML)
- **Phase 3**: Review risks 2, 4 (links, hierarchy)
- **Phase 4**: Review risk 3 (tables)
- **Phase 6**: Review risks 6, 7 (performance, rate limits)
- **Phase 7**: User feedback on all risks

---

## Lessons from Similar Projects

### Confluence Exporters
- **Lesson 1**: Link resolution is always tricky
  - **Applied**: Defer to Phase 3, build comprehensive index
- **Lesson 2**: Tables are a common pain point
  - **Applied**: Test early, document limitations
- **Lesson 3**: Users want both formats (source + readable)
  - **Applied**: Keep HTML, add Markdown

### Markdown Generators
- **Lesson 1**: YAML frontmatter escaping is critical
  - **Applied**: Implement proper escaping from start
- **Lesson 2**: Whitespace consistency matters
  - **Applied**: Normalize blank lines, trim whitespace
- **Lesson 3**: Code blocks need special care
  - **Applied**: HTML escape, preserve language hints

---

**Last Updated**: 2025-11-12  
**Risk Review Status**: Initial assessment complete  
**Next Review**: Before starting Phase 2 integration
