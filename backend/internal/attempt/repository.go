package attempt

import "context"

// Repository хранит попытки.
//
// Update использует оптимистичную блокировку: сохранение проходит, только если
// версия совпадает с сохранённой. Это ровно то, что PostgreSQL-адаптер
// Backend №2 выразит как UPDATE ... WHERE id = $1 AND version = $2, поэтому
// замена хранилища не потребует переписывать сервис.
type Repository interface {
	Create(ctx context.Context, attempt Attempt) error
	Get(ctx context.Context, id ID) (Attempt, error)
	Update(ctx context.Context, attempt Attempt) (Attempt, error)
}
