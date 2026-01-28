package service

import (
	"context"

	"github.com/samandr77/bot-test/internal/models"
	"github.com/samandr77/bot-test/internal/repository"
)

type StateService interface {
	Get(ctx context.Context, userID int64) (*models.UserState, error)
	SetMode(ctx context.Context, userID int64, mode models.UserMode) error
	SetWaitingFor(ctx context.Context, userID int64, waiting models.WaitingFor) error
	ClearWaiting(ctx context.Context, userID int64) error
	SetSoraPrompt(ctx context.Context, userID int64, prompt string) error
	SetSoraImage(ctx context.Context, userID int64, imageURL string) error
	SetSoraDuration(ctx context.Context, userID int64, duration int) error
	SetSoraFormat(ctx context.Context, userID int64, format string) error
	ToggleSoraHD(ctx context.Context, userID int64) (bool, error)
	SetNanoPrompt(ctx context.Context, userID int64, prompt string) error
	SetNanoImage(ctx context.Context, userID int64, imageURL string) error
	ResetSora(ctx context.Context, userID int64) error
	ResetNano(ctx context.Context, userID int64) error
}

type stateService struct {
	repo repository.StateRepository
}

func NewStateService(repo repository.StateRepository) StateService {
	return &stateService{
		repo: repo,
	}
}

func (s *stateService) Get(ctx context.Context, userID int64) (*models.UserState, error) {
	state, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (s *stateService) save(ctx context.Context, userID int64, state *models.UserState) error {
	return s.repo.Save(ctx, userID, state)
}

func (s *stateService) SetMode(ctx context.Context, userID int64, mode models.UserMode) error {
	state, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	state.Mode = mode
	state.WaitingFor = models.WaitingNone
	return s.save(ctx, userID, state)
}

func (s *stateService) SetWaitingFor(ctx context.Context, userID int64, waiting models.WaitingFor) error {
	state, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	state.WaitingFor = waiting
	return s.save(ctx, userID, state)
}

func (s *stateService) ClearWaiting(ctx context.Context, userID int64) error {
	state, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	state.WaitingFor = models.WaitingNone
	return s.save(ctx, userID, state)
}

func (s *stateService) SetSoraPrompt(ctx context.Context, userID int64, prompt string) error {
	state, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	state.Sora.Prompt = prompt
	state.WaitingFor = models.WaitingNone
	return s.save(ctx, userID, state)
}

func (s *stateService) SetSoraImage(ctx context.Context, userID int64, imageURL string) error {
	state, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	state.Sora.ImageURL = imageURL
	state.WaitingFor = models.WaitingNone
	return s.save(ctx, userID, state)
}

func (s *stateService) SetSoraDuration(ctx context.Context, userID int64, duration int) error {
	state, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	state.Sora.Duration = duration
	return s.save(ctx, userID, state)
}

func (s *stateService) SetSoraFormat(ctx context.Context, userID int64, format string) error {
	state, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	state.Sora.Format = format
	return s.save(ctx, userID, state)
}

func (s *stateService) ToggleSoraHD(ctx context.Context, userID int64) (bool, error) {
	state, err := s.Get(ctx, userID)
	if err != nil {
		return false, err
	}
	state.Sora.HD = !state.Sora.HD
	if err := s.save(ctx, userID, state); err != nil {
		return false, err
	}
	return state.Sora.HD, nil
}

func (s *stateService) SetNanoPrompt(ctx context.Context, userID int64, prompt string) error {
	state, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	state.Nano.Prompt = prompt
	state.WaitingFor = models.WaitingNone
	return s.save(ctx, userID, state)
}

func (s *stateService) SetNanoImage(ctx context.Context, userID int64, imageURL string) error {
	state, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	state.Nano.ImageURL = imageURL
	state.WaitingFor = models.WaitingNone
	return s.save(ctx, userID, state)
}

func (s *stateService) ResetSora(ctx context.Context, userID int64) error {
	state, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	state.Sora = models.SoraSettings{
		Duration: 10,
		Format:   "16:9",
		HD:       false,
	}
	return s.save(ctx, userID, state)
}

func (s *stateService) ResetNano(ctx context.Context, userID int64) error {
	state, err := s.Get(ctx, userID)
	if err != nil {
		return err
	}
	state.Nano = models.NanoSettings{}
	return s.save(ctx, userID, state)
}
