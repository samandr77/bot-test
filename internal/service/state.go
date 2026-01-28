package service

import (
	"context"

	"github.com/samandr77/bot-test/internal/models"
	"github.com/samandr77/bot-test/internal/repository"
)

type StateService interface {
	Get(ctx context.Context, userID int64) *models.UserState
	SetMode(ctx context.Context, userID int64, mode models.UserMode)
	SetWaitingFor(ctx context.Context, userID int64, waiting models.WaitingFor)
	ClearWaiting(ctx context.Context, userID int64)
	SetSoraPrompt(ctx context.Context, userID int64, prompt string)
	SetSoraImage(ctx context.Context, userID int64, imageURL string)
	SetSoraDuration(ctx context.Context, userID int64, duration int)
	SetSoraFormat(ctx context.Context, userID int64, format string)
	ToggleSoraHD(ctx context.Context, userID int64) bool
	SetNanoPrompt(ctx context.Context, userID int64, prompt string)
	SetNanoImage(ctx context.Context, userID int64, imageURL string)
	ResetSora(ctx context.Context, userID int64)
	ResetNano(ctx context.Context, userID int64)
}

type stateService struct {
	repo repository.StateRepository
}

func NewStateService(repo repository.StateRepository) StateService {
	return &stateService{
		repo: repo,
	}
}

func (s *stateService) Get(ctx context.Context, userID int64) *models.UserState {
	state, err := s.repo.Get(ctx, userID)
	if err != nil {
		return models.NewUserState()
	}
	return state
}

func (s *stateService) save(ctx context.Context, userID int64, state *models.UserState) error {
	return s.repo.Save(ctx, userID, state)
}

func (s *stateService) SetMode(ctx context.Context, userID int64, mode models.UserMode) {
	state := s.Get(ctx, userID)
	state.Mode = mode
	state.WaitingFor = models.WaitingNone
	s.save(ctx, userID, state)
}

func (s *stateService) SetWaitingFor(ctx context.Context, userID int64, waiting models.WaitingFor) {
	state := s.Get(ctx, userID)
	state.WaitingFor = waiting
	s.save(ctx, userID, state)
}

func (s *stateService) ClearWaiting(ctx context.Context, userID int64) {
	state := s.Get(ctx, userID)
	state.WaitingFor = models.WaitingNone
	s.save(ctx, userID, state)
}

func (s *stateService) SetSoraPrompt(ctx context.Context, userID int64, prompt string) {
	state := s.Get(ctx, userID)
	state.Sora.Prompt = prompt
	state.WaitingFor = models.WaitingNone
	s.save(ctx, userID, state)
}

func (s *stateService) SetSoraImage(ctx context.Context, userID int64, imageURL string) {
	state := s.Get(ctx, userID)
	state.Sora.ImageURL = imageURL
	state.WaitingFor = models.WaitingNone
	s.save(ctx, userID, state)
}

func (s *stateService) SetSoraDuration(ctx context.Context, userID int64, duration int) {
	state := s.Get(ctx, userID)
	state.Sora.Duration = duration
	s.save(ctx, userID, state)
}

func (s *stateService) SetSoraFormat(ctx context.Context, userID int64, format string) {
	state := s.Get(ctx, userID)
	state.Sora.Format = format
	s.save(ctx, userID, state)
}

func (s *stateService) ToggleSoraHD(ctx context.Context, userID int64) bool {
	state := s.Get(ctx, userID)
	state.Sora.HD = !state.Sora.HD
	s.save(ctx, userID, state)
	return state.Sora.HD
}

func (s *stateService) SetNanoPrompt(ctx context.Context, userID int64, prompt string) {
	state := s.Get(ctx, userID)
	state.Nano.Prompt = prompt
	state.WaitingFor = models.WaitingNone
	s.save(ctx, userID, state)
}

func (s *stateService) SetNanoImage(ctx context.Context, userID int64, imageURL string) {
	state := s.Get(ctx, userID)
	state.Nano.ImageURL = imageURL
	state.WaitingFor = models.WaitingNone
	s.save(ctx, userID, state)
}

func (s *stateService) ResetSora(ctx context.Context, userID int64) {
	state := s.Get(ctx, userID)
	state.Sora = models.SoraSettings{
		Duration: 10,
		Format:   "16:9",
		HD:       false,
	}
	s.save(ctx, userID, state)
}

func (s *stateService) ResetNano(ctx context.Context, userID int64) {
	state := s.Get(ctx, userID)
	state.Nano = models.NanoSettings{}
	s.save(ctx, userID, state)
}
