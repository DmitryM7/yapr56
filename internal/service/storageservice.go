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
	ErrNoLuhnNumber          = errors.New("LUHN CHECKSUMM ERROR")
	ErrOrderExists           = errors.New("ORDER WITH NUMBER EXISTS")
	ErrDublicateOrder        = errors.New("DUBLICATE ORDER")
)

type StorageService struct {
	db          *sql.DB
	DatabaseDSN string
}

func (s *StorageService) connect() error {
	db, err := sql.Open("pgx", s.DatabaseDSN)

	if err != nil {
		return fmt.Errorf("CANT do sql.open: [%w]", err)
	}

	if err := db.PingContext(context.Background()); err != nil {
		return fmt.Errorf("CANT PING DB: [%w]", err)
	}

	s.db = db

	return nil
}

func (s *StorageService) splitToDigitRevers(num int) ([]int, error) {
	digitArray := []int{}
	testNum := num

	for testNum > 0 {
		digitArray = append(digitArray, testNum%10)
		testNum = (int)(testNum / 10)
	}

	return digitArray, nil
}

func (s *StorageService) checkByLuhn(num int) error {

	digitArray, err := s.splitToDigitRevers(num)

	if err != nil {
		return err
	}

	cK := 0

	for k, v := range digitArray {

		if (k+1)%2 == 0 {
			testDigit := v * 2
			if testDigit > 9 {
				smallDigitArray, err := s.splitToDigitRevers(testDigit)
				if err != nil {
					return err
				}
				testDigit = smallDigitArray[0] + smallDigitArray[1]
			}
			v = testDigit
		}

		cK += v
	}

	if cK%10 != 0 {
		return ErrNoLuhnNumber
	}

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
	var personID int

	p.Crdt = time.Now()
	p.Updt = p.Crdt

	err := s.db.QueryRowContext(ctx, `INSERT INTO person (login,password,crdt,updt) VALUES($1,$2,$3,$4) RETURNING id`,
		p.Login,
		p.Pass,
		p.Crdt,
		p.Updt).Scan(&personID)

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

func (s *StorageService) CreateOrder(ctx context.Context, p models.Person, order models.POrder) (models.POrder, error) {

	var orderID int

	err := s.checkByLuhn(order.Extnum)

	if err != nil {
		return order, fmt.Errorf("NO VALID LUHN [%w]", err)
	}

	order.Pid = p.GetID()
	order.Crdt = time.Now()
	order.Updt = order.Crdt

	err = s.db.QueryRowContext(ctx, `INSERT INTO porder (pid,extnum,crdt,updt) VALUES($1,$2,$3,$4) RETURNING id`,
		order.Pid,
		order.Extnum,
		order.Crdt,
		order.Updt).Scan(&orderID)

	if err != nil {
		var perr *pgconn.PgError

		if errors.As(err, &perr) && perr.Code != pgerrcode.UniqueViolation {
			return order, err
		}

		dublicateOrder, err := s.GetOrder(ctx, order)

		if err != nil {
			return order, err
		}

		if dublicateOrder.GetPID() == order.Pid {
			return order, ErrDublicateOrder
		} else {
			return order, ErrOrderExists
		}
	}

	return order, nil
}

func (s *StorageService) GetOrder(ctx context.Context, order models.POrder) (models.POrder, error) {

	var status sql.NullString

	err := s.db.QueryRowContext(ctx, "SELECT id,pid,extnum,status,crdt,updt FROM porder WHERE extnum=$1", order.Extnum).
		Scan(&order.ID,
			&order.Pid,
			&order.Extnum,
			&status,
			&order.Crdt,
			&order.Updt)

	order.Status = status.String

	if err != nil {
		return order, err
	}

	return order, nil

}

func (s *StorageService) GetPersonByID(ctx context.Context, id int) (models.Person, error) {
	person := models.Person{}

	row := s.db.QueryRowContext(ctx, "SELECT * FROM person WHERE id=$1", id)

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
			return person, err
		}
		return person, fmt.Errorf("CAN'T SEARCH PERSON BY ID [%w]", err)
	}

	return person, nil
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
