package html

type Topic struct {
	Title   string
	Short   string
	Content string
	Slug    string
}

type Help struct {
	DefaultGeneralTopics []Topic
	OtherGeneralTopics   []Topic
	DefaultExamples      []Topic
	OtherExamples        []Topic
	DefaultApplications  []Topic
	OtherApplications    []Topic
	DefaultTutorials     []Topic
	OtherTutorials       []Topic
}
