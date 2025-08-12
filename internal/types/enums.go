package types

type AnsweredState int

const (
	AnsweredUnknown AnsweredState = iota
	AnsweredTrue
	AnsweredFalse
)
