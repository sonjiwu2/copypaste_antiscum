package memory

import (
	"context"
	"sync"
	"time"

	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/profile"
	"github.com/sonjiwu2/copypaste_antiscum/backend/internal/weeklytest"
)

type WeeklyTestRepository struct {
	mu     sync.RWMutex
	byID   map[weeklytest.ID]weeklytest.Test
	byWeek map[string]weeklytest.ID
}

func NewWeeklyTestRepository() *WeeklyTestRepository {
	return &WeeklyTestRepository{
		byID:   make(map[weeklytest.ID]weeklytest.Test),
		byWeek: make(map[string]weeklytest.ID),
	}
}

func (r *WeeklyTestRepository) FindByWeek(
	ctx context.Context,
	owner profile.ID,
	weekStart time.Time,
) (weeklytest.Test, error) {
	if err := ctx.Err(); err != nil {
		return weeklytest.Test{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.byWeek[weeklyTestWeekKey(owner, weekStart)]
	if !exists {
		return weeklytest.Test{}, weeklytest.ErrNotFound
	}

	return cloneWeeklyTest(r.byID[id]), nil
}

func (r *WeeklyTestRepository) FindLatest(
	ctx context.Context,
	owner profile.ID,
) (weeklytest.Test, error) {
	if err := ctx.Err(); err != nil {
		return weeklytest.Test{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest weeklytest.Test
	found := false
	for _, candidate := range r.byID {
		if candidate.Owner != owner {
			continue
		}
		if !found || candidate.CreatedAt.After(latest.CreatedAt) ||
			(candidate.CreatedAt.Equal(latest.CreatedAt) && candidate.ID > latest.ID) {
			latest = candidate
			found = true
		}
	}
	if !found {
		return weeklytest.Test{}, weeklytest.ErrNotFound
	}

	return cloneWeeklyTest(latest), nil
}

func (r *WeeklyTestRepository) FindByID(ctx context.Context, id weeklytest.ID) (weeklytest.Test, error) {
	if err := ctx.Err(); err != nil {
		return weeklytest.Test{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	found, exists := r.byID[id]
	if !exists {
		return weeklytest.Test{}, weeklytest.ErrNotFound
	}

	return cloneWeeklyTest(found), nil
}

func (r *WeeklyTestRepository) Create(
	ctx context.Context,
	created weeklytest.Test,
) (weeklytest.Test, error) {
	if err := ctx.Err(); err != nil {
		return weeklytest.Test{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := weeklyTestWeekKey(created.Owner, created.WeekStart)
	if existingID, exists := r.byWeek[key]; exists {
		return cloneWeeklyTest(r.byID[existingID]), nil
	}

	stored := cloneWeeklyTest(created)
	r.byID[created.ID] = stored
	r.byWeek[key] = created.ID

	return cloneWeeklyTest(stored), nil
}

func (r *WeeklyTestRepository) SaveSubmission(
	ctx context.Context,
	id weeklytest.ID,
	owner profile.ID,
	submission weeklytest.Submission,
) (weeklytest.Test, error) {
	if err := ctx.Err(); err != nil {
		return weeklytest.Test{}, err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	found, exists := r.byID[id]
	if !exists {
		return weeklytest.Test{}, weeklytest.ErrNotFound
	}
	if found.Owner != owner {
		return weeklytest.Test{}, weeklytest.ErrForbidden
	}
	if found.Submission == nil {
		copied := cloneSubmission(submission)
		found.Submission = &copied
		r.byID[id] = found
	}

	return cloneWeeklyTest(found), nil
}

func weeklyTestWeekKey(owner profile.ID, weekStart time.Time) string {
	return string(owner) + ":" + weekStart.UTC().Format(time.DateOnly)
}

func cloneWeeklyTest(source weeklytest.Test) weeklytest.Test {
	cloned := source
	cloned.Questions = make([]weeklytest.Question, len(source.Questions))
	for index, question := range source.Questions {
		cloned.Questions[index] = question
		cloned.Questions[index].Options = append([]string(nil), question.Options...)
	}
	if source.Submission != nil {
		copied := cloneSubmission(*source.Submission)
		cloned.Submission = &copied
	}

	return cloned
}

func cloneSubmission(source weeklytest.Submission) weeklytest.Submission {
	cloned := source
	cloned.Answers = append([]weeklytest.Answer(nil), source.Answers...)

	return cloned
}
