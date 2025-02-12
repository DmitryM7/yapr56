package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/DmitryM7/yapr56.git/internal/logger"
	"github.com/DmitryM7/yapr56.git/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrUserExists            = errors.New("USER EXISTS")
	ErrUserCredentialInvalid = errors.New("USER CREDENTIAL INVALID")
)

type StorageService struct {
	db          *sql.DB
	DatabaseDSN string
}

func (s *StorageService) connect() error {
	db, err := sql.Open("pgx", s.DatabaseDSN)

	if err != nil {
		return fmt.Errorf("CANT do sql.open: [%v]", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		return fmt.Errorf("CANT PING DB: [%v]", err)
	}

	s.db = db

	return nil
}

func (s *StorageService) GetPesonByCredential(ctx context.Context, login, pass string) (models.Person, error) {
	person := models.Person{}

	row := s.db.QueryRowContext(ctx, "SELECT * FROM person WHERE login=$1 AND password=$2", login, pass)

	var fullname, surname, name, status sql.NullString

	err := row.Scan(&person.ID,
		&person.Login,
		&person.Pass,
		&fullname,
		&surname,
		&name,
		&status,
		&person.Crdt,
		&person.Updt)

	person.Fullname = fullname.String
	person.Surname = surname.String
	person.Name = name.String
	person.Status = status.String

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return person, ErrUserCredentialInvalid
		}
		return person, fmt.Errorf("CAN'T SEARCH PERSON BY CREDENTIAL [%w]", err)
	}

	return person, nil
}

func (s *StorageService) CreatePeson(ctx context.Context, p models.Person) (models.Person, error) {
	var (
		personID     int
		currDateTime time.Time = time.Now()
	)

	err := s.db.QueryRowContext(ctx, `INSERT INTO person (login,password,crdt,updt) VALUES($1,$2,$3,$4) RETURNING id`,
		p.Login,
		p.Pass,
		currDateTime,
		currDateTime).Scan(&personID)

	if err != nil {
		var perr *pgconn.PgError

		if errors.As(err, &perr) && perr.Code == pgerrcode.UniqueViolation {
			return p, ErrUserExists
		}

		return models.Person{}, fmt.Errorf("CAN'T CREATE PERSON [%w]", err)
	}

	p.ID = uint(personID)

	return p, nil
}

func NewStorageService(log logger.Lg, dsn string) (StorageService, error) {
	s := StorageService{
		DatabaseDSN: dsn,
	}

	if err := s.connect(); err != nil {
		return s, fmt.Errorf("CAN'T CONNECT TO DB [%w]", err)
	}

	return s, nil
}
