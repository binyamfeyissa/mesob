package app

import (
	"context"
	"io"
)

type PartnerWebhookUseCase struct {
	Processor WebhookProcessor
}

type PartnerWebhookInput struct {
	Partner   string
	Body      io.Reader
	Signature string
	Timestamp string
}

type PartnerWebhookOutput struct {
	Received bool
}

func (uc *PartnerWebhookUseCase) Execute(ctx context.Context, in PartnerWebhookInput) (PartnerWebhookOutput, error) {
	if uc.Processor == nil {
		// Processor not wired yet — acknowledge receipt without processing
		return PartnerWebhookOutput{Received: true}, nil
	}
	if err := uc.Processor.Process(ctx, in.Partner, in.Body, in.Signature, in.Timestamp); err != nil {
		return PartnerWebhookOutput{}, err
	}
	return PartnerWebhookOutput{Received: true}, nil
}
