package pgs

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"avito-test-task/internal/models/entity"
	"avito-test-task/internal/storage"
	"avito-test-task/internal/storage/pgs/common/bannertag"
)

func (s *Storage) UpdateBanner(ctx context.Context, b *entity.UpdatableBanner) (err error) {
	const comp = "storage.pgs.UpdateBanner"

	tx, err := s.dbPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", comp, err)
	}
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			err = fmt.Errorf("%s: %w", comp, err)
		}
	}()

	t := tx.QueryRow(ctx, "SELECT id FROM banner WHERE id = $1 FOR UPDATE;", b.ID)
	err = t.Scan(&b.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: %w", comp, storage.ErrBannerNotFound)
		}
		return fmt.Errorf("%s: %w", comp, err)
	}

	batch := new(pgx.Batch)

	q, args := buildUpdateBannerQuery(b)
	batch.Queue(q, args...)
	batch.Queue("DELETE FROM banner_tag WHERE banner_id = $1;", b.ID)
	if b.TagIDs != nil {
		batch.Queue(bannertag.InsertTagsQuery(b.ID, *b.TagIDs))
	}

	bres := tx.SendBatch(ctx, batch)
	defer bres.Close()
	_, err = bres.Exec()
	if err != nil {
		return fmt.Errorf("%s: %w", comp, err)
	}
	_ = bres.Close()
	err = tx.Commit(ctx)
	pgErr := new(pgconn.PgError)
	if errors.As(err, &pgErr) && pgErr.Code == "P0001" { // P0001 when trigger is fired
		return fmt.Errorf("%s: %w", comp, storage.ErrBannerAlreadyExists)
	} else if err != nil {
		return fmt.Errorf("%s: %w", comp, err)
	}

	return nil
}

func buildUpdateBannerQuery(b *entity.UpdatableBanner) (string, []any) {
	var (
		args = make([]any, 0)
		sb   strings.Builder
	)

	sb.WriteString("UPDATE banner SET ")

	if b.Title != nil {
		sb.WriteString("title = $")
		sb.WriteString(fmt.Sprintf("%d, ", len(args)+1))
		args = append(args, *b.Title)
	}

	if b.Text != nil {
		sb.WriteString("text = $")
		sb.WriteString(fmt.Sprintf("%d, ", len(args)+1))
		args = append(args, *b.Text)
	}

	if b.FeatureID != nil {
		sb.WriteString("feature_id = $")
		sb.WriteString(fmt.Sprintf("%d, ", len(args)+1))
		args = append(args, *b.FeatureID)
	}

	if b.IsActive != nil {
		sb.WriteString("is_active = $")
		sb.WriteString(fmt.Sprintf("%d, ", len(args)+1))
		args = append(args, *b.IsActive)
	}

	if b.URL != nil {
		sb.WriteString("url = $")
		sb.WriteString(fmt.Sprintf("%d, ", len(args)+1))
		args = append(args, *b.URL)
	}

	sb.WriteString("updated_at = NOW()")

	query := sb.String()

	query += fmt.Sprintf(" WHERE id = %d;", b.ID)

	return query, args
}
