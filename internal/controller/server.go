package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

	Srv struct {
		Log        logger.Lg
		Service    service.Service
		JwtService IJwtService
	}
)

func (s *Srv) actMiddleWare(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
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
		s.Log.Warnln("EMPTY BODY")
		return
	}

	request := UserRegisterRequest{}

	err = json.Unmarshal(body, &request)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.Log.Errorln("CAN'T UNMARSHAL USER REQUEST IN REGISTER ACTION:", err)
		return
	}

	person := models.Person{
		Login: request.Login,
		Pass:  request.Password,
	}

	person, err = s.Service.CreatePeson(person)

	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			w.WriteHeader(http.StatusConflict)
			s.Log.Debugln(fmt.Sprintf("LOGIN [%s] IS BUSY", person.Login))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		s.Log.Warnln("CAN'T GET PERSON BY CREDENTIAL:", err)
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

	person, err := s.Service.GetPesonByCredential(p.Login, p.Password)

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

}

func (s *Srv) actOrders(w http.ResponseWriter, r *http.Request) {

}

func (s *Srv) actAcctBalance(w http.ResponseWriter, r *http.Request) {

}

func (s *Srv) actAcctDebit(w http.ResponseWriter, r *http.Request) {

}

func (s *Srv) actAcctStatement(w http.ResponseWriter, r *http.Request) {

}
func NewServer(log logger.Lg, serv service.Service, jwt IJwtService) (*Srv, error) {

	return &Srv{
		Log:        log,
		Service:    serv,
		JwtService: jwt,
	}, nil

}
