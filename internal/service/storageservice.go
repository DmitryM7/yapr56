package service

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/DmitryM7/yapr56.git/internal/logger"
	"github.com/DmitryM7/yapr56.git/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

var (
	ErrUserExists            = errors.New("USER EXISTS")
	ErrUserCredentialInvalid = errors.New("USER CREDENTIAL INVALID")
	ErrNoLuhnNumber          = errors.New("LUHN CHECKSUMM ERROR")
	ErrOrderExists           = errors.New("ORDER WITH NUMBER EXISTS")
	ErrDublicateOrder        = errors.New("DUBLICATE ORDER")
	ErrNoFixedBalance        = errors.New("NO FIXED BALANCE YET")
	ErrRedSaldo              = errors.New("RED SALDO")
	//go:embed migrations/*.sql
	embedMigrations embed.FS
)

type StorageService struct {
	db          *sql.DB
	DatabaseDSN string
}

const (
	StatusNew        = "NEW"
	StatusProcessing = "PROCESSING"
	Invalid          = "INVALID"
	Processed        = "PROCESSED"
)

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
	var personID, acctID, acctSerial int

	p.Crdt = time.Now()
	p.Updt = p.Crdt

	tx, err := s.db.Begin()

	if err != nil {
		return p, fmt.Errorf("CAN'T OPEN TRANSACT: [%v]", err)
	}

	err = tx.QueryRowContext(ctx, `INSERT INTO person (login,password,crdt,updt) VALUES($1,$2,$3,$4) RETURNING id`,
		p.Login,
		p.Pass,
		p.Crdt,
		p.Updt).Scan(&personID)

	if err != nil {
		var perr *pgconn.PgError

		if errors.As(err, &perr) && perr.Code == pgerrcode.UniqueViolation {
			tx.Rollback()
			return p, ErrUserExists
		}

		tx.Rollback()

		return models.Person{}, fmt.Errorf("CAN'T CREATE PERSON [%w]", err)
	}

	err = tx.QueryRowContext(ctx, `SELECT nextval('acctserial')`).Scan(&acctSerial)

	if err != nil {
		tx.Rollback()
		return models.Person{}, fmt.Errorf("CAN'T ACCT SEQUENCE VALUE [%w]", err)
	}
	fmt.Println("SEQ:", acctSerial)
	err = tx.QueryRowContext(ctx, `INSERT INTO acct (acct,person,sign,crdt,updt) VALUES($1,$2,'П',$3,$4) RETURNING id`,
		`408178101`+fmt.Sprintf("%011d", acctSerial),
		personID,
		p.Crdt,
		p.Updt).Scan(&acctID)

	if err != nil {
		var perr *pgconn.PgError

		if errors.As(err, &perr) && perr.Code == pgerrcode.UniqueViolation {
			tx.Rollback()
			return p, ErrUserExists
		}

		tx.Rollback()

		return models.Person{}, fmt.Errorf("CAN'T CREATE PERSON [%w]", err)
	}

	p.ID = uint(personID)

	tx.Commit()

	return p, nil
}

func (s *StorageService) CreateOrder(ctx context.Context, p models.Person, order models.POrder) (models.POrder, error) {

	var orderID int

	err := s.checkByLuhn(order.Extnum)

	if err != nil {
		return order, fmt.Errorf("NO VALID LUHN [%w]", err)
	}

	order.Status = StatusNew

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

func (s *StorageService) GetOrders(ctx context.Context, p models.Person) ([]models.POrder, error) {

	var status sql.NullString

	result := []models.POrder{}

	rows, err := s.db.QueryContext(ctx, `SELECT id,pid,extnum,status,accrual,crdt,updt 
										 FROM porder 
										 WHERE pid=$1 
										ORDER BY crdt DESC`, p.GetID())

	if err != nil {
		return result, fmt.Errorf("CAN'T GET ORDERS [%w]", err)
	}

	order := models.POrder{}

	for rows.Next() {

		err := rows.Scan(&order.ID,
			&order.Pid,
			&order.Extnum,
			&status,
			&order.Accrual,
			&order.Crdt,
			&order.Updt)
		order.Status = status.String

		if err != nil {
			return result, err
		}

		result = append(result, order)

	}
	return result, nil
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

func (s *StorageService) getMoveByDb(ctx context.Context, acct string, opdate time.Time) ([]models.Opentry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT opentry.id,
    											opentry.person,
    											opentry.porder,
    											opentry.status,
    											opentry.opdate,
    											opentry.acctdb,
    											opentry.acctcr,
    											opentry.sum1,
    											opentry.sum2,    
    											opentry.crdt,
    											opentry.updt,
												porder.extnum
	FROM  opentry
	LEFT JOIN porder ON porder.id=opentry.porder
	WHERE acctdb=$1 
	AND opdate>=$2`,
		acct,
		opdate)

	if err != nil {
		return nil, fmt.Errorf("CAN'T READ OPENTRY BY DB: [%v]", err)
	}

	res := make([]models.Opentry, 10)
	var status sql.NullString
	var extNum sql.NullInt32

	for rows.Next() {
		opentry := models.Opentry{}
		err := rows.Scan(
			&opentry.ID,
			&opentry.Person,
			&opentry.Porder,
			&status,
			&opentry.Opdate,
			&opentry.Acctdb,
			&opentry.Acctcr,
			&opentry.Sum1,
			&opentry.Sum2,
			&opentry.Crdt,
			&opentry.Updt,
			&extNum)

		opentry.Status = status.String
		opentry.OrderExtNum = int(extNum.Int32)

		if err != nil {
			return nil, fmt.Errorf("CAN'T READ OPENTRY BY DB: [%v]", err)
		}

		res = append(res, opentry)

	}

	return res, nil
}

func (s *StorageService) getMoveByCr(ctx context.Context, acct string, opdate time.Time) ([]models.Opentry, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id,
    											person,
    											porder,
    											status,
    											opdate,
    											acctdb,
    											acctcr,
    											sum1,
    											sum2,    
    											crdt,
    											updt
	FROM  opentry
	WHERE acctcr=$1 
	AND opdate>=$2`,
		acct,
		opdate)

	if err != nil {
		return nil, fmt.Errorf("CAN'T READ OPENTRY BY CR: [%v]", err)
	}

	res := make([]models.Opentry, 10)
	var status sql.NullString

	for rows.Next() {
		opentry := models.Opentry{}
		err := rows.Scan(
			&opentry.ID,
			&opentry.Person,
			&opentry.Porder,
			&status,
			&opentry.Opdate,
			&opentry.Acctdb,
			&opentry.Acctcr,
			&opentry.Sum1,
			&opentry.Sum2,
			&opentry.Crdt,
			&opentry.Updt)

		opentry.Status = status.String

		if err != nil {
			return nil, fmt.Errorf("CAN'T READ OPENTRY BY CR: [%v]", err)
		}

		res = append(res, opentry)

	}

	return res, nil
}

func (s *StorageService) getLastFixBalance(ctx context.Context, acct models.Acct) (models.AcctBal, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id,
											 person,
											 opdate,
											 acct,
											 balance,
											 db,
											 cr,
											 crdt,
											 updt
	                                  FROM acct 
									  WHERE acct=$1 AND opdate>$2
									  ORDER BY opdate DESC LIMIT 1`,
		acct.Acct,
		acct.Crdt)

	acctbal := models.AcctBal{}

	err := row.Scan(
		&acctbal.ID,
		&acctbal.Person,
		&acctbal.Opdate,
		&acctbal.Acct,
		&acctbal.Balance,
		&acctbal.Db,
		&acctbal.Cr,
		&acctbal.Crdt,
		&acctbal.Updt)

	if err != nil {
		if err != sql.ErrNoRows {
			return models.AcctBal{}, fmt.Errorf("CAN'T READ ACCTBAL INFO [%v]", err)
		}

		return models.AcctBal{}, ErrNoFixedBalance
	}

	return acctbal, nil

}
func (s *StorageService) calcBalanceByAcct(ctx context.Context, acct models.Acct) (int, error) {

	balance := 0

	acctbal, err := s.getLastFixBalance(ctx, acct)

	if err != nil {
		if !errors.Is(err, ErrNoFixedBalance) {
			return 0, fmt.Errorf("CAN'T READ ACCTBAL INFO [%v]", err)
		}
	} else {
		balance += acctbal.Balance
	}

	rows, err := s.getMoveByDb(ctx, acct.Acct, acctbal.Opdate)

	if err != nil {
		return 0, fmt.Errorf("CAN'T READ OPENTRY BY CR INFO [%v]", err)
	}

	for _, opentry := range rows {
		if acct.Sign == "А" {
			balance += opentry.Sum1
		} else if acct.Sign == "П" {
			balance -= opentry.Sum1
		}
	}

	rows, err = s.getMoveByCr(ctx, acct.Acct, acctbal.Opdate)

	if err != nil {
		return 0, fmt.Errorf("CAN'T READ OPENTRY BY DB INFO [%v]", err)
	}

	for _, opentry := range rows {

		if acct.Sign == "А" {
			balance -= opentry.Sum1
		} else if acct.Sign == "П" {
			balance += opentry.Sum1
		}

	}

	return balance, nil
}

func (s *StorageService) getPersonAccts(ctx context.Context, p models.Person) ([]models.Acct, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT * FROM acct WHERE person=$1", p.GetID())

	if err != nil {
		return nil, fmt.Errorf("CAN'T FIND PERSON acct [%v]", err)
	}

	res := make([]models.Acct, 5)

	for rows.Next() {
		var status, sign sql.NullString
		acct := models.Acct{}
		err := rows.Scan(&acct.ID, &acct.Acct, &acct.Person, &acct.Crdt, &acct.Updt)
		acct.Status = status.String
		acct.Sign = sign.String

		if err != nil {
			return nil, fmt.Errorf("CAN'T CREATE ACCT STRUCT [%v]", err)
		}

		res = append(res, acct)
	}

	return res, nil

}

func (s *StorageService) GetBalance(ctx context.Context, p models.Person) (int, error) {

	b := 0

	accts, err := s.getPersonAccts(ctx, p)

	if err != nil {
		return 0, fmt.Errorf("CAN'T FIND PERSON ACCT [%v]", err)
	}

	for _, acct := range accts {

		if b0, err := s.calcBalanceByAcct(ctx, acct); err == nil {
			b += b0
		}
	}
	return b, nil
}

func (s *StorageService) Getwithdrawn(ctx context.Context, p models.Person) (int, error) {
	b := 0

	accts, err := s.getPersonAccts(ctx, p)

	if err != nil {
		return 0, fmt.Errorf("CAN'T GET PERSON ACCTS")
	}

	for _, acct := range accts {
		fixedBalance, err := s.getLastFixBalance(ctx, acct)

		if err != nil {
			if !errors.Is(err, ErrNoFixedBalance) {
				return 0, fmt.Errorf("CAN'T GET LAST FIXED BALANCE: [%v]", err)
			}
		}

		if acct.Sign == "П" {
			b += fixedBalance.Db
			rows, err := s.getMoveByDb(ctx, acct.Acct, fixedBalance.Opdate)

			if err != nil {
				return 0, fmt.Errorf("CAN'T GET ACCT MOBY BY DB: [%v]", err)
			}
			for _, opentry := range rows {
				b += opentry.Sum1
			}
		} else if acct.Sign == "А" {
			b += fixedBalance.Cr

			rows, err := s.getMoveByCr(ctx, acct.Acct, fixedBalance.Opdate)

			if err != nil {
				return 0, fmt.Errorf("CAN'T GET ACCT MOBY BY DB: [%v]", err)
			}
			for _, opentry := range rows {
				b += opentry.Sum1
			}
		}

	}

	return b, nil

}

func (s *StorageService) CreateWithdrawn(ctx context.Context, p models.Person, o models.POrder, sum int) (models.Opentry, error) {
	balance, err := s.GetBalance(ctx, p)

	if err != nil {
		return models.Opentry{}, fmt.Errorf("CAN GET BALANCE: [%v]", err)
	}

	if sum > balance {
		return models.Opentry{}, ErrRedSaldo
	}

	accts, err := s.getPersonAccts(ctx, p)

	if err != nil {
		return models.Opentry{}, fmt.Errorf("CAN'T GET PERSON ACCT: [%v]", err)
	}

	if len(accts) == 0 {
		return models.Opentry{}, fmt.Errorf("NO ACTIVE ACCT FOR PERSON")
	}

	acct := accts[0]

	opentry := models.Opentry{
		Person:      p.GetID(),
		Porder:      o.ID,
		OrderExtNum: o.Extnum,
		Opdate:      time.Now(),
		Sum1:        sum,
		Crdt:        time.Now(),
		Updt:        time.Now(),
	}

	if acct.Sign == "П" {
		opentry.Acctdb = acct.Acct
		opentry.Acctcr = "30102810000000000001"
	} else if acct.Sign == "А" {
		opentry.Acctcr = acct.Acct
		opentry.Acctdb = "30102810000000000001"
	}

	var opentryId uint
	err = s.db.QueryRowContext(ctx, `INSERT INTO opentry (person,porder,orderextnum,opdate,acctdb,acctcr,sum1,crdt,updt) 
	                                 VALUES($1,$2,$3,$4) RETURNING id`,
		opentry.Person,
		opentry.Porder,
		opentry.OrderExtNum,
		opentry.Status,
		opentry.Opdate,
		opentry.Acctdb,
		opentry.Acctcr,
		opentry.Sum1,
		opentry.Crdt,
		opentry.Updt).
		Scan(&opentryId)

	if err != nil {
		return models.Opentry{}, fmt.Errorf("CAN'T INSERT OPENTRY")
	}

	opentry.ID = opentryId

	return opentry, nil
}
func (s *StorageService) GetWithdrawals(ctx context.Context, p models.Person) ([]models.Opentry, error) {
	accts, err := s.getPersonAccts(ctx, p)

	if err != nil {
		return nil, fmt.Errorf("CAN'T FIND PERSON ACCT [%v]", err)
	}

	rows := make([]models.Opentry, 10)

	for _, acct := range accts {
		if acct.Sign == "П" {
			r, e := s.getMoveByDb(ctx, acct.Acct, acct.Crdt)

			if e != nil {
				return nil, fmt.Errorf("ACCT IS PASSIVE AND CAN'T GET MOVE BY DB [%v]", err)
			}

			rows = append(rows, r...)
		} else if acct.Sign == "А" {
			r, e := s.getMoveByCr(ctx, acct.Acct, acct.Crdt)

			if e != nil {
				return nil, fmt.Errorf("ACCT IS ACTIVE AND CAN'T GET MOVE BY DB [%v]", err)
			}

			rows = append(rows, r...)
		}
	}

	return rows, nil

}

func (s *StorageService) RunMigrations() error {

	goose.SetBaseFS(embedMigrations)
	goose.SetDialect("postgres")

	if err := goose.Up(s.db, "migrations"); err != nil {
		return err
	}
	return nil
}

func NewStorageService(log logger.Lg, dsn string) (StorageService, error) {
	s := StorageService{
		DatabaseDSN: dsn,
	}

	if err := s.connect(); err != nil {
		return s, fmt.Errorf("CAN'T CONNECT TO DB [%w]", err)
	}

	err := s.RunMigrations()

	if err != nil {
		log.Panicln(err)
	}

	return s, nil
}
