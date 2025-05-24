package types

type AnsweredState int

const (
	AnsweredUnknow AnsweredState = iota
	AnsweredTrue
	AnsweredFalse
)
