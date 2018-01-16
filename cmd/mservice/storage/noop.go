package storage

import "context"

type NoopStorage struct {
}

func (s *NoopStorage) MakeCrew(ctx context.Context, pid string) error {
	return nil
}

func (s *NoopStorage) RemCrew(ctx context.Context, pid string) error {
	return nil
}

func (s *NoopStorage) GetCrew(ctx context.Context, pid string) ([]*MachineState, error) {
	return nil, nil
}

func (s *NoopStorage) WriteState(ctx context.Context, pid string, ss []*MachineState) error {
	return nil
}

func (s *NoopStorage) Open(ctx context.Context) error {
	return nil
}

func (s *NoopStorage) Close(ctx context.Context) error {
	return nil
}
