# Documentation Index

## 🎯 Start Here

**For Current Work**: [`STATE.md`](../STATE.md) (project root)  
**For Reference**: [`CRUSH.md`](../CRUSH.md) (agent guide)  
**For Users**: [`README.md`](../README.md) (user documentation)

## 📖 Documentation Structure

### Primary (Root Level)
- **STATE.md** - 🔴 **START HERE** - Current project state, next actions, decisions
- **README.md** - User guide (installation, usage, features)
- **CRUSH.md** - Agent reference (commands, patterns, conventions)

### Development Documentation (docs/)

#### Workflow & Guidelines
- **AGENT.md** - Agent workflow and state tracking methodology
- **README.md** - This file (documentation index)

#### Technical Documentation
- **architecture.md** - Technical design for markdown export feature
- **CONFLUENCE_FORMATS.md** - HTML pattern analysis (479 pages analyzed)
- **MARKDOWN_EXPORT_PLAN.md** - Original 7-phase development plan

#### Progress & Completion Reports
- **PHASE_2_COMPLETE.md** - Phase 2 implementation details and summary
- **SESSION_SUMMARY.md** - Latest session recap
- **SESSION_COMPLETE.md** - Session completion notes

#### Historical (docs/archive/)
- **goals.md** - Original feature goals and success criteria
- **progress.md** - Detailed phase tracking (superseded by STATE.md)
- **learnings.md** - Key insights and decisions (superseded by STATE.md Decision Log)
- **implementation_plan.md** - Step-by-step plan (superseded by STATE.md Next Actions)
- **testing_plan.md** - Test strategy
- **assumptions_and_risks.md** - Risk assessment
- **archive/README.md** - Information about archived files

## 📊 Documentation Methodology

### State Tracking (New as of 2025-11-12)

**Single Source of Truth**: `STATE.md`

**Update Triggers**:
- After making key decisions → Decision Log
- After completing work → Current Focus + Next Actions  
- When blocked → Blockers section
- When changing tasks → Current Focus

**Benefits**:
- Reduced cognitive load (one file to check)
- Fast context loading for agents
- Clear temporal progression (timestamped decisions)
- Actionable next steps (not vague goals)

### Reference Documentation

**CRUSH.md**: Commands, patterns, conventions (changes infrequently)

**docs/**: Deep technical details, research, historical records

## 🔄 Migration Notes

**Previous Approach** (before 2025-11-12):
- 7 scattered state files (goals, progress, learnings, implementation_plan, testing_plan, assumptions_and_risks, plus SESSION_SUMMARY)
- Overlapping information across multiple files
- Hard to know which file was current
- Higher cognitive load for context switching

**New Approach** (as of 2025-11-12):
- Single `STATE.md` file for living state
- `CRUSH.md` for stable reference material
- Historical files archived to `docs/archive/`
- Clear update triggers and methodology

## 📚 Quick Links

### For AI Agents
1. Start: [`STATE.md`](../STATE.md) - Current focus and next actions
2. Reference: [`CRUSH.md`](../CRUSH.md) - Commands and patterns
3. Deep dive: `docs/` files as needed

### For Developers
1. User docs: [`README.md`](../README.md)
2. Current state: [`STATE.md`](../STATE.md)
3. Technical design: [`architecture.md`](architecture.md)
4. Research: [`CONFLUENCE_FORMATS.md`](CONFLUENCE_FORMATS.md)

### For Context on Decisions
- Recent decisions: [`STATE.md`](../STATE.md) Decision Log
- Historical decisions: [`archive/learnings.md`](archive/learnings.md)
- Phase completion: [`PHASE_2_COMPLETE.md`](PHASE_2_COMPLETE.md)

## 🔧 How to Use This Documentation

### Session Start (AI Agents)
1. Read `STATE.md` first (Current Focus + Next 3 Actions)
2. Check Blockers section (any human input needed?)
3. Review recent Decision Log entries
4. Consult `CRUSH.md` for commands/patterns as needed
5. Deep dive into `docs/` only if specific context needed

### Session Start (Human Developers)
1. Check `STATE.md` for current status
2. Review Next 3 Actions for what to work on
3. Read recent Decision Log for context
4. Refer to `CRUSH.md` for commands/patterns

### Making Decisions
1. Consider alternatives and trade-offs
2. Add to `STATE.md` Decision Log with timestamp
3. Include: Decision, Why, Implementation, Trade-off, Outcome
4. Update Current Focus if changing direction

### Completing Work
1. Update `STATE.md` Current Focus with new status
2. Update Next 3 Actions (remove completed, add new)
3. Commit code changes with descriptive message
4. Update `CRUSH.md` if new patterns emerged

## 📋 Document Purposes

### STATE.md (Living Document)
**Purpose**: Single source of truth for "what's happening now"

**Contains**:
- Current Focus (what's being worked on)
- Next 3 Actions (immediate priorities)
- Context Snapshot (minimal info to resume)
- Blockers (active impediments)
- Decision Log (recent key decisions with rationale)
- Quick Context Links (pointers to details)

**Update**: After every significant decision or task completion

### CRUSH.md (Reference Guide)
**Purpose**: Comprehensive agent reference material

**Contains**:
- Essential commands (build, test, run)
- Code patterns and conventions
- Architecture overview
- Important gotchas
- Testing patterns
- Common development tasks
- Performance characteristics

**Update**: When patterns change or new patterns emerge (infrequently)

### docs/ Files (Deep Context)
**Purpose**: Detailed technical documentation and historical records

**Contains**:
- Technical architecture (`architecture.md`)
- Research artifacts (`CONFLUENCE_FORMATS.md`)
- Phase completion reports (`PHASE_2_COMPLETE.md`)
- Original plans (`MARKDOWN_EXPORT_PLAN.md`)
- Historical state tracking (`archive/`)

**Update**: When documenting research, completing phases, or recording history

## 🎓 Best Practices

### For Context Loading
✅ Progressive disclosure: STATE.md → CRUSH.md → docs/  
✅ Start with current state, expand as needed  
✅ Use links to navigate between documents

### For State Recording
✅ Be specific: "Add ConvertWithMetadata() to converter.go"  
✅ Be actionable: "Run ./test-export.sh with credentials"  
✅ Be current: Update before pausing work

### For Decision Logging
✅ Capture WHY not just WHAT  
✅ Include alternatives considered  
✅ Timestamp for temporal reasoning  
✅ Keep concise (5-8 lines)

### For Documentation Maintenance
✅ STATE.md: Update frequently (after every decision/task)  
✅ CRUSH.md: Update occasionally (new patterns only)  
✅ docs/: Update for research/phase completion

---

**Last Updated**: 2025-11-12  
**Methodology**: State Stack Pattern (see AGENT.md)  
**Primary State File**: [`STATE.md`](../STATE.md)
