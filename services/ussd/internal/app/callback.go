package app

import (
	"context"
	"github.com/mesob-wallet/ussd/internal/domain"
)

type CallbackInput struct {
	SessionID   string
	MSISDN      string
	Input       string
	ServiceCode string
}

type CallbackOutput struct {
	Message  string `json:"message"`
	Continue bool   `json:"continue"`
}

type CallbackUseCase struct {
	Sessions SessionStore
}

func (uc *CallbackUseCase) Execute(ctx context.Context, in CallbackInput) (*CallbackOutput, error) {
	var sess *domain.Session
	if uc.Sessions != nil {
		var err error
		sess, err = uc.Sessions.Get(ctx, in.SessionID)
		if err != nil {
			sess = nil
		}
	}
	if sess == nil {
		sess = &domain.Session{
			ID:     in.SessionID,
			MSISDN: in.MSISDN,
			State:  domain.StateMain,
			Lang:   "am",
		}
	}
	msg, cont := sess.Next(in.Input)
	if uc.Sessions != nil {
		if cont {
			uc.Sessions.Save(ctx, sess)
		} else {
			uc.Sessions.Delete(ctx, in.SessionID)
		}
	}
	return &CallbackOutput{Message: msg, Continue: cont}, nil
}
