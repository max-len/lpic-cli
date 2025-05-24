package types

type Testset struct {
	TestsetID          string `bson:"testsetId"`
	TestsetName        string `bson:"testsetName"`
	TestsetDescription string `bson:"testsetDescription"`
	QuestionsIds       []int  `bson:"questionsIds" json:"QuestionsIds"`
}
