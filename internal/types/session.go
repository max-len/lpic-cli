package types

type CertificationSession struct {
	QuestionIndex        map[string]int `bson:"currentQuestionIndex"`
	CurrentQuestionIndex int            `bson:"currentQuestionIndex"`
	Testset              *Testset
}

func NewCertificationSession(Testset *Testset) *CertificationSession {
	return &CertificationSession{
		QuestionIndex:        make(map[string]int),
		CurrentQuestionIndex: 0,
		Testset:              Testset,
	}
}

func (s *CertificationSession) GetCurrentQuestionIndex() int {
	return s.CurrentQuestionIndex
}

func (s *CertificationSession) GetFirst() int {
	s.CurrentQuestionIndex = 0
	return s.CurrentQuestionIndex
}

func (s *CertificationSession) GetAndIncIndex() int {
	if s.CurrentQuestionIndex >= len(s.Testset.QuestionsIds) {
		s.CurrentQuestionIndex = len(s.Testset.QuestionsIds) - 1
		return s.CurrentQuestionIndex
	}

	s.CurrentQuestionIndex++
	return s.CurrentQuestionIndex
}

func (s *CertificationSession) GetAndDecIndex() int {
	if s.CurrentQuestionIndex <= 0 {
		s.CurrentQuestionIndex = 0
		return 0
	}
	s.CurrentQuestionIndex--
	return s.CurrentQuestionIndex
}
