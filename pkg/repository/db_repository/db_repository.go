package db_repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/vpnvsk/amunetip-patent-upload/internal/config"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"log/slog"
	"strings"
)

const batchSize = 500

type DBRepository struct {
	db  *sqlx.DB
	log *slog.Logger
	cfg *config.Config
}

func NewDBRepository(db *sqlx.DB, log *slog.Logger, cfg *config.Config) *DBRepository {
	return &DBRepository{
		db:  db,
		log: log,
		cfg: cfg,
	}
}

func (r *DBRepository) SavePatents(ctx context.Context, patents []model.FilteredFullPatent, transactionId, bundleId uuid.UUID) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	for i := 0; i < len(patents); i += 500 {
		end := i + batchSize
		if end > len(patents) {
			end = len(patents)
		}
		if err := r.insertPatentsBulk(ctx, patents[i:end], tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert batch patents failed: %w", err)
		}
		if err := r.insertInventorsBulk(ctx, patents[i:end], tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert batch inventors failed: %w", err)
		}
		if err := r.insertInventorPatentLinksBulk(ctx, patents[i:end], tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert batch patentsinventors failed: %w", err)
		}
		if err := r.insertAssigneesBulk(ctx, patents[i:end], tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert batch assignee failed: %w", err)
		}
		if err := r.insertAssigneePatentLinksBulk(ctx, patents[i:end], tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert batch patentsassignee failed: %w", err)
		}
		if err := r.insertJurisdictionsBulk(ctx, patents[i:end], tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert batch jur failed: %w", err)
		}
		if err := r.insertJurisdictionsPatentLinksBulk(ctx, patents[i:end], tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert batch patentsjur failed: %w", err)
		}
		if err := r.insertPatentTransactionLinkBulk(ctx, patents[i:end], transactionId, tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert batch transactionpat failed: %w", err)
		}
		if err := r.insertPatentBundleLinkBulk(ctx, patents[i:end], bundleId, tx); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert batch bundlepatents failed: %w", err)
		}

	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit failed: %w", err)
	}
	return err
}

func (r *DBRepository) insertPatentsBulk(ctx context.Context, patents []model.FilteredFullPatent, tx *sqlx.Tx) error {
	const fieldsPerRow = 9

	placeholders := make([]string, 0, len(patents))
	args := make([]interface{}, 0, len(patents)*fieldsPerRow)

	for i, p := range patents {
		offset := i * fieldsPerRow
		placeholders = append(placeholders, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d) ",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8, offset+9,
		))
		args = append(args,
			p.ID,
			p.Patent.Title,
			p.Description,
			p.Abstract,
			p.Patent.PublicationNumber,
			p.Patent.EarliestPriorityDate,
			p.Patent.EstimatedExpiryDate,
			p.Patent.ApplicationDate,
			p.Patent.SimpleLegalStatus,
		)
	}

	query := fmt.Sprintf(`
        INSERT INTO patent (
            id, title, description, abstract,
            publication_number, earliest_priority_date,
            estimated_expiry_date, application_date, simple_legal_status
        )
        VALUES %s`, strings.Join(placeholders, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *DBRepository) insertInventorsBulk(ctx context.Context, patents []model.FilteredFullPatent, tx *sqlx.Tx) error {
	var (
		args         []interface{}
		placeholders []string
		idx          = 1
	)

	for _, p := range patents {
		for _, name := range p.Patent.InventorsNames {
			placeholders = append(placeholders, fmt.Sprintf("($%d)", idx))
			args = append(args, name)
			idx++
		}
	}

	if len(placeholders) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
        INSERT INTO inventor (full_name)
        VALUES %s
        ON CONFLICT DO NOTHING`, strings.Join(placeholders, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *DBRepository) insertInventorPatentLinksBulk(ctx context.Context, patents []model.FilteredFullPatent, tx *sqlx.Tx) error {
	placeholders := make([]string, 0, len(patents))
	args := make([]interface{}, 0, len(patents)*2)
	idx := 1
	for _, p := range patents {
		for _, inventor := range p.Patent.InventorsNames {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", idx, idx+1))
			args = append(args, p.ID, inventor)
			idx += 2
		}
	}
	if len(placeholders) == 0 {
		return nil
	}
	query := fmt.Sprintf(`
INSERT INTO patentinventorlink (patent_id, inventor_name) VALUES %s`, strings.Join(placeholders, ","))
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *DBRepository) insertAssigneesBulk(ctx context.Context, patents []model.FilteredFullPatent, tx *sqlx.Tx) error {
	var (
		args         []interface{}
		placeholders []string
		idx          = 1
	)

	for _, p := range patents {
		for _, name := range p.Patent.Assignee {
			placeholders = append(placeholders, fmt.Sprintf("($%d)", idx))
			args = append(args, name)
			idx++
		}
	}

	if len(placeholders) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
        INSERT INTO standardizedcurrentassignee (name)
        VALUES %s
        ON CONFLICT DO NOTHING`, strings.Join(placeholders, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *DBRepository) insertAssigneePatentLinksBulk(ctx context.Context, patents []model.FilteredFullPatent, tx *sqlx.Tx) error {
	placeholders := make([]string, 0, len(patents))
	args := make([]interface{}, 0, len(patents)*2)
	idx := 1
	for _, p := range patents {
		for _, assignee := range p.Patent.Assignee {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", idx, idx+1))
			args = append(args, p.ID, assignee)
			idx += 2
		}
	}
	if len(placeholders) == 0 {
		return nil
	}
	query := fmt.Sprintf(`
INSERT INTO patentstandardizedcurrentassigneelink (patent_id, standardized_current_assignee_name) VALUES %s`, strings.Join(placeholders, ","))
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *DBRepository) insertJurisdictionsBulk(ctx context.Context, patents []model.FilteredFullPatent, tx *sqlx.Tx) error {
	var (
		args         []interface{}
		placeholders []string
		idx          = 1
	)

	for _, p := range patents {
		for _, name := range p.Patent.SimpleFamilyJurisdiction {
			placeholders = append(placeholders, fmt.Sprintf("($%d)", idx))
			args = append(args, name)
			idx++
		}
	}

	if len(placeholders) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
        INSERT INTO simplefamilyjurisdiction (name)
        VALUES %s
        ON CONFLICT DO NOTHING`, strings.Join(placeholders, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *DBRepository) insertJurisdictionsPatentLinksBulk(ctx context.Context, patents []model.FilteredFullPatent, tx *sqlx.Tx) error {
	placeholders := make([]string, 0, len(patents))
	args := make([]interface{}, 0, len(patents)*2)
	idx := 1
	for _, p := range patents {
		for _, jur := range p.Patent.SimpleFamilyJurisdiction {
			placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", idx, idx+1))
			args = append(args, p.ID, jur)
			idx += 2
		}
	}
	if len(placeholders) == 0 {
		return nil
	}
	query := fmt.Sprintf(`
INSERT INTO patentsimplefamilyjurisdictionlink (patent_id, family_jurisdiction_name) VALUES %s`, strings.Join(placeholders, ","))
	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *DBRepository) insertPatentTransactionLinkBulk(ctx context.Context, patents []model.FilteredFullPatent, transactionId uuid.UUID, tx *sqlx.Tx) error {
	placeholders := make([]string, 0, len(patents))
	args := make([]interface{}, 0, len(patents)*2)

	for i, p := range patents {
		idx := i*2 + 1
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", idx, idx+1))
		args = append(args, p.ID, transactionId)
	}

	if len(placeholders) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
        INSERT INTO patenttransactionlink (patent_id, transaction_id)
        VALUES %s`, strings.Join(placeholders, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *DBRepository) insertPatentBundleLinkBulk(ctx context.Context, patents []model.FilteredFullPatent, bundleId uuid.UUID, tx *sqlx.Tx) error {
	placeholders := make([]string, 0, len(patents))
	args := make([]interface{}, 0, len(patents)*2)

	for i, p := range patents {
		idx := i*2 + 1
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", idx, idx+1))
		args = append(args, p.ID, bundleId)
	}

	if len(placeholders) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
        INSERT INTO bundlepatentlink (patent_id, bundle_id)
        VALUES %s`, strings.Join(placeholders, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *DBRepository) saveClaims(claims *model.Claim, tx *sqlx.Tx) error {
	return nil
}
