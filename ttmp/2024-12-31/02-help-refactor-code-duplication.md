### 4. Code Duplication

The codebase contains several patterns of unnecessary duplication:

#### A. Repeated Section Query Methods

Each section type has nearly identical query methods:
```go
func (s *Section) DefaultGeneralTopic() []*Section {
    return NewSectionQuery().
        ReturnTopics().
        ReturnOnlyTopics(s.Slug).
        ReturnOnlyShownByDefault().
        FilterSections(s).
        FindSections(s.HelpSystem.Sections)
}

func (s *Section) DefaultExamples() []*Section {
    return NewSectionQuery().
        ReturnExamples().           // Only this line differs
        ReturnOnlyTopics(s.Slug).
        ReturnOnlyShownByDefault().
        FilterSections(s).
        FindSections(s.HelpSystem.Sections)
}

func (s *Section) DefaultApplications() []*Section {
    return NewSectionQuery().
        ReturnApplications().       // Only this line differs
        ReturnOnlyTopics(s.Slug).
        ReturnOnlyShownByDefault().
        FilterSections(s).
        FindSections(s.HelpSystem.Sections)
}
```

#### B. Redundant Default/Other Pattern

Every section type implements the same Default/Other pattern:
```go
// In NewHelpPage
for _, section := range sections {
    switch section.SectionType {
    case SectionGeneralTopic:
        if section.ShowPerDefault {
            ret.DefaultGeneralTopics = append(ret.DefaultGeneralTopics, section)
        } else {
            ret.OtherGeneralTopics = append(ret.OtherGeneralTopics, section)
        }
        ret.AllGeneralTopics = append(ret.DefaultGeneralTopics, ret.OtherGeneralTopics...)
    case SectionExample:
        if section.ShowPerDefault {
            ret.DefaultExamples = append(ret.DefaultExamples, section)
        } else {
            ret.OtherExamples = append(ret.OtherExamples, section)
        }
        ret.AllExamples = append(ret.DefaultExamples, ret.OtherExamples...)
    // ... same pattern repeats for Applications and Tutorials
    }
}
```

#### C. Similar Validation Methods

Multiple similar validation methods that could be generalized:
```go
func (s *Section) IsForCommand(command string) bool {
    return strings2.StringInSlice(command, s.Commands)
}

func (s *Section) IsForFlag(flag string) bool {
    return strings2.StringInSlice(flag, s.Flags)
}

func (s *Section) IsForTopic(topic string) bool {
    return strings2.StringInSlice(topic, s.Topics)
}
```

### Proposed Solutions for Duplication

1. **Generic Query Builder**
```go
type SectionQueryBuilder[T SectionType] struct {
    filters []QueryFilter
}

func (s *Section) QueryByType[T SectionType](sType T) []*Section {
    return NewSectionQuery[T]().
        WithType(sType).
        WithTopics(s.Slug).
        WithDefaultVisibility(true).
        Execute(s.HelpSystem.Sections)
}
```

2. **Generic Collection Handler**
```go
type SectionCollection[T SectionType] struct {
    Default []T
    Other   []T
    All     []T
}

func (sc *SectionCollection[T]) Add(item T, isDefault bool) {
    if isDefault {
        sc.Default = append(sc.Default, item)
    } else {
        sc.Other = append(sc.Other, item)
    }
    sc.All = append(sc.All, item)
}
```

3. **Generic Validation**
```go
type StringMatcher interface {
    GetStrings() []string
}

func IsMatchingString(needle string, haystack StringMatcher) bool {
    return strings2.StringInSlice(needle, haystack.GetStrings())
}
```