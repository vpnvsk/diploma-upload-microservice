package db_repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/vpnvsk/amunetip-patent-upload/internal/config"
	"github.com/vpnvsk/amunetip-patent-upload/internal/model"
	"log/slog"
)

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

func (r *DBRepository) SavePatents(data *model.ParsedPatentDataDB) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	err = r.savePatents(data.Patents, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	return err
}

func (r *DBRepository) savePatents(patents []*model.ParsedPatent, tx *sqlx.Tx) error {
	return nil
}

func (r *DBRepository) saveInventors(inventors *[]model.Inventor, tx *sqlx.Tx) error {
	return nil
}

func (r *DBRepository) saveInventorPatentLinks(links *[]model.PatentInventorLink, tx *sqlx.Tx) error {
	return nil
}

func (r *DBRepository) saveAssignees(assignees *[]model.StandardizedCurrentAssignee, tx *sqlx.Tx) error {
	return nil
}

func (r *DBRepository) saveAssigneePatentLinks(links *[]model.PatentStandardizedCurrentAssigneeLink, tx *sqlx.Tx) error {
	return nil
}

func (r *DBRepository) saveJurisdictions(jurisdictions *[]model.SimpleFamilyJurisdiction, tx *sqlx.Tx) error {
	return nil
}

func (r *DBRepository) saveJurisdictionsPatentLinks(links *[]model.PatentSimpleFamilyJurisdictionLink, tx *sqlx.Tx) error {
	return nil
}

func (r *DBRepository) saveClaims(claims *model.Claim, tx *sqlx.Tx) error {
	return nil
}
