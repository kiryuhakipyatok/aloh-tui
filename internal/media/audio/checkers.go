package audio

type Checkers interface {
	CheckUserIsSpeaking() bool
}

func (ae *audioEngine) CheckUserIsSpeaking() bool {
	return ae.userIsSpeaking.Load()
}