package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/DmitryM7/yapr56.git/internal/logger"
	"github.com/DmitryM7/yapr56.git/internal/models"
	"github.com/DmitryM7/yapr56.git/internal/service"
)

type (
	IJwtService interface {
		GetJwtStr(uid int) (string, error)
		UnloadUserIDJwt(tokenString string) (int, error)
		TokenExpired() time.Duration
	}

	IStorage interface {
		GetPesonByCredential(ctx context.Context, login, pass string) (models.Person, error)
		CreatePeson(ctx context.Context, p models.Person) (models.Person, error)
		CreateOrder(ctx context.Context, p models.Person, order models.POrder) (models.POrder, error)
		GetOrder(ctx context.Context, order models.POrder) (models.POrder, error)
		GetPersonByID(ctx context.Context, id int) (models.Person, error)
		GetOrders(ctx context.Context, p models.Person) ([]models.POrder, error)
		GetBalance(ctx context.Context, p models.Person) (int, error)
		Getwithdrawn(ctx context.Context, p models.Person) (int, error)
		GetWithdrawals(ctx context.Context, p models.Person) ([]models.Opentry, error)
	}

	Srv struct {
		Log           logger.Lg
		Service       IStorage
		JwtService    IJwtService
		NoAuthActions map[string]string
	}

	contextParam string
)

func (s *Srv) actMiddleWare(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		s.Log.Debugln("URL PATH IS:", r.URL.Path)

		if _, e := s.NoAuthActions[r.URL.Path]; !e {
			cookie, err := r.Cookie("token")

			if err != nil {
				s.Log.Debugln("CAN'T READ COOKIE:", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			personId, err := s.JwtService.UnloadUserIDJwt(cookie.Value)

			if err != nil {
				s.Log.Debugln("CAN'T UNLOAD ID FROM JWT:", err)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(ctx, contextParam("CurrPersonID"), personId)

		}

		next.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(f)
}
func (s *Srv) actUserRegister(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorln("CAN'T READ BODY:", err)
		return
	}

	defer func() {
		err := r.Body.Close()
		if err != nil {
			s.Log.Warnln("CAN'T CLOSE BODY")
		}
	}()

	if string(body) == "" {
		w.WriteHeader(http.StatusBadRequest)
		s.Log.Debugln("EMPTY BODY")
		return
	}

	request := UserRegisterRequest{}

	err = json.Unmarshal(body, &request)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.Log.Warnln("CAN'T UNMARSHAL USER REQUEST IN REGISTER ACTION:", err)
		return
	}

	person := models.Person{
		Login: request.Login,
		Pass:  request.Password,
	}

	ctx := context.TODO()

	person, err = s.Service.CreatePeson(ctx, person)

	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			w.WriteHeader(http.StatusConflict)
			s.Log.Debugln(fmt.Sprintf("LOGIN [%s] IS BUSY", person.Login))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Warnln("CAN'T CREATE PERSON BY CREDENTIAL:", err)
		return
	}

	jwtToken, err := s.JwtService.GetJwtStr(int(person.ID))

	if err != nil {
		s.Log.Errorln("CAN'T CREATE JWT FOR USER:", person.ID)
		return
	}

	s.Log.Debugln(fmt.Sprintf("PERSON WAS CREATE id=%d,login=%s", person.ID, person.Login))

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   jwtToken,
		Expires: time.Now().Add(s.JwtService.TokenExpired() * time.Minute),
	})

	w.WriteHeader(http.StatusOK)
}
func (s *Srv) actUserLogin(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Warnln("CAN'T READ BODY")
		return
	}

	defer func() {
		err := r.Body.Close()
		if err != nil {
			s.Log.Warnln("CAN'T CLOSE BODY")
		}
	}()

	p := UserAuthRequest{}

	err = json.Unmarshal(body, &p)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.Log.Errorln("CAN'T UNMARSHAL BODY")
		return
	}

	ctx := context.TODO()
	person, err := s.Service.GetPesonByCredential(ctx, p.Login, p.Password)

	if err != nil {
		if errors.Is(err, service.ErrUserCredentialInvalid) {
			w.WriteHeader(http.StatusUnauthorized)
			s.Log.Infoln("INVALID USER NAME OR PASS:", p.Login)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorln("CAN'T GET USER BY LOGIN AND PASS:", err)
		return
	}

	jwtToken, err := s.JwtService.GetJwtStr(int(person.ID))

	if err != nil {
		s.Log.Errorln("CAN'T CREATE JWT FOR USER:", person.ID)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   jwtToken,
		Expires: time.Now().Add(s.JwtService.TokenExpired() * time.Minute),
	})

	w.WriteHeader(http.StatusOK)
}
func (s *Srv) actOrdersUpload(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Warnln("CAN'T READ BODY")
		return
	}

	defer func() {
		err := r.Body.Close()
		if err != nil {
			s.Log.Warnln("CAN'T CLOSE BODY")
		}
	}()

	if string(body) == "" {
		w.WriteHeader(http.StatusBadRequest)
		s.Log.Infoln("EMPTY BODY")
		return
	}

	extNum, err := strconv.Atoi(string(body))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.Log.Infoln("CAN'T PARSE ORDER NUMBER TO INT")
		return
	}

	ctx := r.Context()

	if currPersonId, ok := ctx.Value(contextParam("CurrPersonID")).(int); ok {
		currPerson, err := s.Service.GetPersonByID(ctx, currPersonId)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			s.Log.Infoln("CAN'T FIND PERSON WITH ID=", currPersonId)
			return
		}

		order := models.POrder{
			Extnum: extNum,
		}

		order, err = s.Service.CreateOrder(ctx, currPerson, order)

		if err != nil {

			if errors.Is(err, service.ErrNoLuhnNumber) {
				w.WriteHeader(http.StatusUnprocessableEntity)
				s.Log.Errorln("INCORRECT ORDER NUMBER:", err)
				return
			}

			if errors.Is(err, service.ErrDublicateOrder) {
				w.WriteHeader(http.StatusOK)
				s.Log.Errorln(err)
				return
			}

			if errors.Is(err, service.ErrOrderExists) {
				w.WriteHeader(http.StatusConflict)
				s.Log.Errorln(err)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Errorln(err)
			return

		}

		w.WriteHeader(http.StatusOK)

	}

}

func (s *Srv) actOrders(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if currPersonId, ok := ctx.Value(contextParam("CurrPersonID")).(int); ok {
		person := models.Person{
			ID: uint(currPersonId),
		}

		orders, err := s.Service.GetOrders(ctx, person)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Warnln("CAN'T GET ORDER LIST:", err)
			return
		}

		if len(orders) == 0 {
			w.WriteHeader(http.StatusNoContent)
			s.Log.Infoln("ORDER LIST EMPTY")
			return
		}

		result, err := json.Marshal(orders)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.Log.Warnln("CAN'T MARSHAL ORDER LIST", err)
			return
		}

		s.Log.Infoln(orders)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(result)

		if err != nil {
			s.Log.Warnln("CAN'T WRITE BODY IN ORDER LIST", err)
			return
		}

	}

}

func (s *Srv) actAcctBalance(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	currPersonId, ok := ctx.Value(contextParam("CurrPersonID")).(int)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		s.Log.Warnln("INVALID PERSON ID")
		return
	}

	person, err := s.Service.GetPersonByID(ctx, currPersonId)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		s.Log.Warnln("INVALID PERSON ID")
	}

	balance, err := s.Service.GetBalance(ctx, person)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorln("CAN'T GET BALANCE BY PERSON:[%v]", err)
		return
	}

	withdrawn, err := s.Service.Getwithdrawn(ctx, person)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorln("CAN'T WITHDRAWN BY PERSON:[%v]", err)
		return
	}

	output, err := json.Marshal(BalanceResponce{
		Current:   float32(balance),
		Withdrawn: float32(withdrawn),
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorln("CAN'T MARSHAL DATA:[%v]", err)
		return
	}

	_, err = w.Write(output)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorln("CAN'T WRITE DATA TO BODY:[%v]", err)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func (s *Srv) actAcctDebit(w http.ResponseWriter, r *http.Request) {

}

func (s *Srv) actAcctStatement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	currPersonId, ok := ctx.Value(contextParam("CurrPersonID")).(int)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		s.Log.Warnln("INVALID PERSON ID")
		return
	}

	person, err := s.Service.GetPersonByID(ctx, currPersonId)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		s.Log.Warnln("INVALID PERSON ID")
	}

	rows, err := s.Service.GetWithdrawals(ctx, person)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Warnln("CAN'T GET STATMENT: [%v]", err)
		return

	}

	if len(rows) == 0 {
		w.WriteHeader(http.StatusNoContent)
		s.Log.Warnln("STATMENT IS EMPTY")
		return
	}

	res := make([]WithdrawalsResponce, 10)

	for _, opentry := range rows {
		wr := WithdrawalsResponce{
			Order:       opentry.Porder,
			Sum:         opentry.Sum1,
			ProcessedAt: opentry.Crdt,
		}
		res = append(res, wr)
	}

	output, err := json.Marshal(res)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Warnln("CAN'T MARSHAL RESPONCE: [%v]", err)
		return
	}

	_, err = w.Write(output)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Errorln("CAN'T WRITE DATA TO BODY:[%v]", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func NewServer(log logger.Lg,
	serv IStorage,
	jwt IJwtService) (*Srv, error) {

	var NoAuthActions = map[string]string{
		"/api/user/register": "/api/user/register",
		"/api/user/login":    "/api/user/login",
	}
	return &Srv{
		Log:           log,
		Service:       serv,
		JwtService:    jwt,
		NoAuthActions: NoAuthActions,
	}, nil
}
