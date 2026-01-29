package service

import (
	"context"
	"fmt"

	"github.com/samandr77/bot-test/internal/models"
	"github.com/samandr77/bot-test/internal/pkg/text"
	"github.com/samandr77/bot-test/internal/repository"
)

type StateService interface {
	Get(ctx context.Context, userID int64) (*models.UserState, error)
	GetNextVariant(ctx context.Context, userID int64, key string) (string, error)
	Update(ctx context.Context, userID int64, state *models.UserState) error
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
	state, getErr := s.repo.Get(ctx, userID)
	if getErr != nil {
		return nil, fmt.Errorf("failed to get user state from repo: %w", getErr)
	}
	return state, nil
}

func (s *stateService) GetNextVariant(ctx context.Context, userID int64, key string) (string, error) {
	state, getErr := s.Get(ctx, userID)
	if getErr != nil {
		return "", fmt.Errorf("failed to get state for variant: %w", getErr)
	}

	lastUsed := ""
	if state.LastResponseKeys != nil {
		lastUsed = state.LastResponseKeys[key]
	}

	variant := text.PickRandom(key, lastUsed)

	if state.LastResponseKeys == nil {
		state.LastResponseKeys = make(map[string]string)
	}
	state.LastResponseKeys[key] = variant

	if updateErr := s.Update(ctx, userID, state); updateErr != nil {
		return variant, fmt.Errorf("failed to update state with variant: %w", updateErr)
	}

	return variant, nil
}

func (s *stateService) Update(ctx context.Context, userID int64, state *models.UserState) error {
	saveErr := s.repo.Save(ctx, userID, state)
	if saveErr != nil {
		return fmt.Errorf("failed to save state in repo: %w", saveErr)
	}
	return nil
}

func (s *stateService) SetMode(ctx context.Context, userID int64, mode models.UserMode) error {
	state, getErr := s.Get(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("failed to get state to set mode: %w", getErr)
	}
	state.Mode = mode
	state.WaitingFor = models.WaitingNone
	return s.Update(ctx, userID, state)
}

func (s *stateService) SetWaitingFor(ctx context.Context, userID int64, waiting models.WaitingFor) error {
	state, getErr := s.Get(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("failed to get state to set waiting for: %w", getErr)
	}
	state.WaitingFor = waiting
	return s.Update(ctx, userID, state)
}

func (s *stateService) ClearWaiting(ctx context.Context, userID int64) error {
	state, getErr := s.Get(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("failed to get state to clear waiting: %w", getErr)
	}
	state.WaitingFor = models.WaitingNone
	return s.Update(ctx, userID, state)
}

func (s *stateService) SetSoraPrompt(ctx context.Context, userID int64, prompt string) error {
	state, getErr := s.Get(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("failed to get state to set sora prompt: %w", getErr)
	}
	state.Sora.Prompt = prompt
	state.WaitingFor = models.WaitingNone
	return s.Update(ctx, userID, state)
}

func (s *stateService) SetSoraImage(ctx context.Context, userID int64, imageURL string) error {
	state, getErr := s.Get(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("failed to get state to set sora image: %w", getErr)
	}
	state.Sora.ImageURL = imageURL
	state.WaitingFor = models.WaitingNone
	return s.Update(ctx, userID, state)
}

func (s *stateService) SetSoraDuration(ctx context.Context, userID int64, duration int) error {
	state, getErr := s.Get(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("failed to get state to set sora duration: %w", getErr)
	}
	state.Sora.Duration = duration
	return s.Update(ctx, userID, state)
}

func (s *stateService) SetSoraFormat(ctx context.Context, userID int64, format string) error {
	state, getErr := s.Get(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("failed to get state to set sora format: %w", getErr)
	}
	state.Sora.Format = format
	return s.Update(ctx, userID, state)
}

func (s *stateService) ToggleSoraHD(ctx context.Context, userID int64) (bool, error) {
	state, getErr := s.Get(ctx, userID)
	if getErr != nil {
		return false, fmt.Errorf("failed to get state to toggle sora hd: %w", getErr)
	}
	state.Sora.HD = !state.Sora.HD
	if saveErr := s.Update(ctx, userID, state); saveErr != nil {
		return false, fmt.Errorf("failed to save toggle hd: %w", saveErr)
	}
	return state.Sora.HD, nil
}

func (s *stateService) SetNanoPrompt(ctx context.Context, userID int64, prompt string) error {
	state, getErr := s.Get(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("failed to get state to set nano prompt: %w", getErr)
	}
	state.Nano.Prompt = prompt
	state.WaitingFor = models.WaitingNone
	return s.Update(ctx, userID, state)
}

func (s *stateService) SetNanoImage(ctx context.Context, userID int64, imageURL string) error {
	state, getErr := s.Get(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("failed to get state to set nano image: %w", getErr)
	}
	state.Nano.ImageURL = imageURL
	state.WaitingFor = models.WaitingNone
	return s.Update(ctx, userID, state)
}

func (s *stateService) ResetSora(ctx context.Context, userID int64) error {
	state, getErr := s.Get(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("failed to get state to reset sora: %w", getErr)
	}
	state.Sora = models.SoraSettings{
		Duration: models.DefaultSoraDuration,
		Format:   models.DefaultSoraFormat,
		HD:       false,
	}
	return s.Update(ctx, userID, state)
}

func (s *stateService) ResetNano(ctx context.Context, userID int64) error {
	state, getErr := s.Get(ctx, userID)
	if getErr != nil {
		return fmt.Errorf("failed to get state to reset nano: %w", getErr)
	}
	state.Nano = models.NanoSettings{}
	return s.Update(ctx, userID, state)
}
