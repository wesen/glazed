---
Title: Help Package Section Refactoring
Type: DesignDoc
---

# Current State Analysis

The help package's `Section` type currently suffers from several architectural issues that make the code harder to maintain and evolve:

## Issues

### 1. Bloated Section Type

The `Section` struct handles too many responsibilities:
```go
type Section struct {
    // Content-related fields
    Slug        string
    Title       string
    SubTitle    string
    Short       string
    Content     string

    // Metadata/Classification
    SectionType SectionType
    Topics      []string
    Flags       []string
    Commands    []string

    // Behavior flags
    IsTopLevel      bool
    IsTemplate      bool
    ShowPerDefault  bool
    Order          int

    // Problematic circular reference
    HelpSystem     *HelpSystem `yaml:"_"`
}
```

### 2. Circular Dependencies

The Section type has a circular dependency with HelpSystem:
- Section methods need HelpSystem to query other sections
- HelpSystem contains Sections
- Each Section points back to HelpSystem

### 3. Mixed Responsibilities

Methods on Section handle multiple concerns:
- Content management
- Querying
- Validation
- Rendering

## Refactoring Plan

### Phase 1: Split Section Type

1. Create new types to separate concerns:
```go
type SectionContent struct {
    Slug     string
    Title    string
    SubTitle string
    Short    string
    Content  string
}

type SectionMetadata struct {
    Type    SectionType
    Topics  []string
    Flags   []string
    Commands []string
}

type SectionConfig struct {
    IsTopLevel     bool
    IsTemplate     bool
    ShowPerDefault bool
    Order         int
}

type Section struct {
    Content  SectionContent
    Metadata SectionMetadata
    Config   SectionConfig
}
```

2. Create interfaces for different capabilities:
```go
type ContentProvider interface {
    GetContent() string
    GetTitle() string
    // etc...
}

type Searchable interface {
    MatchesTopic(topic string) bool
    MatchesCommand(command string) bool
    // etc...
}
```

### Phase 2: Repository Pattern

1. Implement a repository to handle section storage and querying:
```go
type SectionRepository interface {
    FindByTopic(topic string) []*Section
    FindByCommand(command string) []*Section
    FindByType(sectionType SectionType) []*Section
    // etc...
}
```

2. Move query logic from Section methods to dedicated query types:
```go
type SectionQuery struct {
    repository *SectionRepository
    filters    []SectionFilter
}

type SectionFilter interface {
    Apply([]*Section) []*Section
}
```

### Phase 3: Break Circular Dependencies

1. Remove HelpSystem reference from Section
2. Pass necessary dependencies through method parameters
3. Use context or dependency injection where appropriate

### Phase 4: Improve Error Handling

1. Create proper error types:
```go
type SectionError struct {
    Code    SectionErrorCode
    Message string
    Cause   error
}

type SectionErrorCode int

const (
    ErrNotFound SectionErrorCode = iota
    ErrInvalidMetadata
    ErrLoadFailed
)
```

### Implementation Order

1. Create new types and interfaces
2. Implement repository pattern
3. Create new query system
4. Migrate existing code to new structure
5. Remove circular dependencies
6. Add improved error handling
7. Update tests
8. Update documentation